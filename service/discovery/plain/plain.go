package plain

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"errors"
	"github.com/gkarlik/quark-go/service"
	"github.com/gkarlik/quark-go/service/discovery"
)

const (
	RegisterServiceURL   = "/register"
	UnregisterServiceURL = "/unregister"
	ListServicesURL      = "/services"
)

type ServerInfo struct {
	Address string   `json:"address"`
	Name    string   `json:"name"`
	Tags    []string `json:"tags"`
	Version string   `json:"version"`
}

type ServiceDiscovery struct {
	mu      *sync.Mutex
	client  *http.Client
	address string
	catalog map[string]*list.List
}

func NewServiceDiscovery(address string) *ServiceDiscovery {
	addr := strings.TrimSuffix(address, "/")

	return &ServiceDiscovery{
		mu: &sync.Mutex{},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		catalog: make(map[string]*list.List),
		address: addr,
	}
}

func (sd *ServiceDiscovery) sendRequest(address string, method string, info service.Info) ([]byte, int, error) {
	si := &ServerInfo{
		Address: info.Address.String(),
		Name:    info.Name,
		Tags:    info.Tags,
		Version: info.Version,
	}

	data, err := json.Marshal(si)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req, err := http.NewRequest(method, address, bytes.NewBuffer(data))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sd.client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return body, resp.StatusCode, nil
}

func (sd *ServiceDiscovery) RegisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	addr := fmt.Sprintf("%s%s", sd.address, RegisterServiceURL)
	_, _, err := sd.sendRequest(addr, http.MethodPost, opts.Info)
	if err != nil {
		return err
	}
	return nil
}

func (sd *ServiceDiscovery) DeregisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	addr := fmt.Sprintf("%s%s", sd.address, UnregisterServiceURL)
	_, _, err := sd.sendRequest(addr, http.MethodPost, opts.Info)
	if err != nil {
		return err
	}
	return nil
}

func (sd *ServiceDiscovery) GetServiceAddress(options ...discovery.Option) (*url.URL, error) {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	addr := fmt.Sprintf("%s%s", sd.address, ListServicesURL)
	data, _, err := sd.sendRequest(addr, http.MethodPost, opts.Info)
	if err != nil {
		return nil, err
	}

	var infos []ServerInfo
	err = json.Unmarshal(data, infos)
	if err != nil {
		return nil, errors.New(http.StatusText(http.StatusInternalServerError))
	}

	var urls []*url.URL
	for _, info := range infos {
		url, _ := url.Parse(info.Address)
		urls = append(urls, url)
	}

	if len(urls) == 0 {
		return nil, nil
	}

	if opts.Strategy == nil {
		return urls[0], nil
	}
	return opts.Strategy.PickServiceAddress(urls)
}

func (sd *ServiceDiscovery) Dispose() {
	if sd.client != nil {
		sd.client = nil
	}
}

func (sd *ServiceDiscovery) Serve(address string) error {
	http.HandleFunc(RegisterServiceURL, sd.registerHandler)
	http.HandleFunc(UnregisterServiceURL, sd.unregisterHandler)
	http.HandleFunc(ListServicesURL, sd.listServicesHandler)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		return err
	}
	return nil
}

func (sd *ServiceDiscovery) decodeServiceInfo(r *http.Request) (*ServerInfo, error) {
	decoder := json.NewDecoder(r.Body)

	var si *ServerInfo
	err := decoder.Decode(si)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return si, nil
}

func (sd *ServiceDiscovery) compareServiceTags(a *ServerInfo, b *ServerInfo) bool {
	if len(a.Tags) == 0 && len(b.Tags) == 0 {
		return true
	}

	counter := 0
	for _, at := range a.Tags {
		for _, bt := range b.Tags {
			if at == bt {
				counter++
			}
		}
	}
	if counter == len(a.Tags) {
		return true
	}
	return false
}

func (sd *ServiceDiscovery) findByServiceInfo(si *ServerInfo) []ServerInfo {
	var result []ServerInfo

	infos, ok := sd.catalog[si.Name]
	if ok {
		for e := infos.Front(); e != nil; e = e.Next() {
			val := e.Value.(ServerInfo)

			if si.Version == val.Version && sd.compareServiceTags(si, &val) {
				result = append(result, val)
			}
		}
	}
	return result
}

func (sd *ServiceDiscovery) deleteByServiceInfo(si *ServerInfo) {
	infos, ok := sd.catalog[si.Name]
	if ok {
		var toDelete []*list.Element
		for e := infos.Front(); e != nil; e = e.Next() {
			val := e.Value.(ServerInfo)

			if si.Version == val.Version && sd.compareServiceTags(si, &val) {
				toDelete = append(toDelete, e)
			}
		}
		sd.mu.Lock()
		for _, d := range toDelete {
			infos.Remove(d)
		}
		sd.mu.Unlock()
	}
}

func (sd *ServiceDiscovery) registerHandler(w http.ResponseWriter, r *http.Request) {
	si, err := sd.decodeServiceInfo(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	infos := sd.findByServiceInfo(si)
	if len(infos) == 0 {
		sd.mu.Lock()
		srvs := list.New()
		srvs.PushBack(*si)
		sd.catalog[si.Name] = srvs
		sd.mu.Unlock()
	}
	w.WriteHeader(http.StatusOK)
}

func (sd *ServiceDiscovery) unregisterHandler(w http.ResponseWriter, r *http.Request) {
	si, err := sd.decodeServiceInfo(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	sd.deleteByServiceInfo(si)

	w.WriteHeader(http.StatusOK)
}

func (sd *ServiceDiscovery) listServicesHandler(w http.ResponseWriter, r *http.Request) {
	si, err := sd.decodeServiceInfo(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}
	infos := sd.findByServiceInfo(si)
	data, err := json.Marshal(infos)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
