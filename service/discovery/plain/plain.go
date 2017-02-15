package plain

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gkarlik/quark-go/logger"
	"github.com/gkarlik/quark-go/service"
	"github.com/gkarlik/quark-go/service/discovery"
)

const (
	// RegisterServiceURL is endpoint url for service registration.
	RegisterServiceURL = "/register"
	// UnregisterServiceURL is endpoint url for service deregistration.
	UnregisterServiceURL = "/unregister"
	// ListServicesURL is endpoint url for registered services list.
	ListServicesURL = "/services"

	componentName = "PlainDiscoveryService"
)

// ServiceInfo represents information about service to be registered in service discovery catalog.
type ServiceInfo struct {
	Address string   `json:"address"` // service address
	Name    string   `json:"name"`    // service name
	Tags    []string `json:"tags"`    // service tags
	Version string   `json:"version"` // service version
}

func (si ServiceInfo) includeTags(tags []string) bool {
	if len(si.Tags) == 0 && len(tags) == 0 {
		return true
	}

	counter := 0
	for _, tag := range si.Tags {
		for _, t := range tags {
			if tag == t {
				counter++
				continue
			}
		}
	}
	if counter == len(si.Tags) {
		return true
	}
	return false
}

func (si ServiceInfo) hasSameTags(tags []string) bool {
	if len(si.Tags) != len(tags) {
		return false
	}

	sort.Strings(si.Tags)
	sort.Strings(tags)

	for i, tag := range si.Tags {
		if tag != tags[i] {
			return false
		}
	}
	return true
}

// ServiceDiscovery represents plain, in-memory, client-server service discovery mechanism.
type ServiceDiscovery struct {
	mu      *sync.Mutex
	client  *http.Client
	address string
	catalog map[string]*list.List
	ln      net.Listener
}

// NewServiceDiscovery creates plain, in-memory, client-server service registration and localization mechanism.
func NewServiceDiscovery(address string) *ServiceDiscovery {
	addr := strings.TrimSuffix(address, "/")

	return &ServiceDiscovery{
		mu: &sync.Mutex{},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		catalog: make(map[string]*list.List),
		address: addr,
		ln:      nil,
	}
}

func (sd *ServiceDiscovery) sendRequest(address string, method string, info service.Info) ([]byte, int, error) {
	si := &ServiceInfo{
		Address: "",
		Name:    info.Name,
		Tags:    info.Tags,
		Version: info.Version,
	}

	if info.Address != nil {
		si.Address = info.Address.String()
	}

	data, err := json.Marshal(si)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"info":      si,
			"component": componentName,
		}, "Cannot convert service info to JSON")
		return nil, http.StatusInternalServerError, err
	}

	req, err := http.NewRequest(method, address, bytes.NewBuffer(data))
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"method":    method,
			"address":   address,
			"data":      data,
			"component": componentName,
		}, "Cannot prepare HTTP request")
		return nil, http.StatusInternalServerError, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sd.client.Do(req)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"request":   req,
			"component": componentName,
		}, "Cannot process HTTP request")
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"response":  resp,
			"component": componentName,
		}, "Cannot read HTTP response body")
		return nil, http.StatusInternalServerError, err
	}
	return body, resp.StatusCode, nil
}

// RegisterService registers service in service discovery catalog.
func (sd *ServiceDiscovery) RegisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	addr := fmt.Sprintf("%s%s", sd.address, RegisterServiceURL)
	_, _, err := sd.sendRequest(addr, http.MethodPost, opts.Info)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"info":      opts.Info,
			"component": componentName,
		}, "Cannot register service")
		return err
	}
	return nil
}

// DeregisterService unregisters service in service discovery catalog.
func (sd *ServiceDiscovery) DeregisterService(options ...discovery.Option) error {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	addr := fmt.Sprintf("%s%s", sd.address, UnregisterServiceURL)
	_, _, err := sd.sendRequest(addr, http.MethodPost, opts.Info)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"info":      opts.Info,
			"component": componentName,
		}, "Cannot unregister service")
		return err
	}
	return nil
}

// GetServiceAddress gets service address from service discovery catalog.
func (sd *ServiceDiscovery) GetServiceAddress(options ...discovery.Option) (*url.URL, error) {
	opts := new(discovery.Options)
	for _, o := range options {
		o(opts)
	}

	addr := fmt.Sprintf("%s%s", sd.address, ListServicesURL)
	data, _, err := sd.sendRequest(addr, http.MethodPost, opts.Info)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"info":      opts.Info,
			"component": componentName,
		}, "Cannot get list of services")
		return nil, err
	}

	var infos []ServiceInfo
	err = json.Unmarshal(data, infos)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot convert JSON string to service info array")
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

// Dispose cleans up ServiceDiscovery instance.
func (sd *ServiceDiscovery) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing service discovery component")

	if sd.client != nil {
		sd.client = nil
	}
	sd.Stop()
}

// Serve starts service discovery HTTP host.
func (sd *ServiceDiscovery) Serve() error {
	http.HandleFunc(RegisterServiceURL, sd.registerHandler)
	http.HandleFunc(UnregisterServiceURL, sd.unregisterHandler)
	http.HandleFunc(ListServicesURL, sd.listServicesHandler)

	ln, err := net.Listen("tcp", sd.address)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"address":   sd.address,
			"component": componentName,
		}, "Cannot listen on address")
		return err
	}

	go func() {
		http.Serve(ln, nil)
	}()
	return nil
}

// Stop stops service discovery HTTP host.
func (sd *ServiceDiscovery) Stop() {
	go func() {
		if sd.ln != nil {
			sd.ln.Close()
		}
	}()
}

func (sd *ServiceDiscovery) decodeServiceInfo(r http.Request) (*ServiceInfo, error) {
	decoder := json.NewDecoder(r.Body)

	var si ServiceInfo
	err := decoder.Decode(&si)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot decode service info from HTTP request")
		return nil, err
	}
	defer r.Body.Close()

	return &si, nil
}

func (sd *ServiceDiscovery) findByServiceInfo(si ServiceInfo) []ServiceInfo {
	// must use make here to return [] instead of null
	result := make([]ServiceInfo, 0)

	infos, ok := sd.catalog[si.Name]
	if ok {
		for e := infos.Front(); e != nil; e = e.Next() {
			val := e.Value.(ServiceInfo)

			if si.Version == val.Version && si.includeTags(val.Tags) {
				result = append(result, val)
			}
		}
	}
	return result
}

func (sd *ServiceDiscovery) findExactByServiceInfo(si ServiceInfo) []ServiceInfo {
	// must use make here to return [] instead of null
	result := make([]ServiceInfo, 0)

	infos, ok := sd.catalog[si.Name]
	if ok {
		for e := infos.Front(); e != nil; e = e.Next() {
			val := e.Value.(ServiceInfo)

			if si.Version == val.Version && si.hasSameTags(val.Tags) {
				result = append(result, val)
			}
		}
	}
	return result
}

func (sd *ServiceDiscovery) deleteByServiceInfo(si ServiceInfo) {
	infos, ok := sd.catalog[si.Name]
	if ok {
		var toDelete []*list.Element
		for e := infos.Front(); e != nil; e = e.Next() {
			val := e.Value.(ServiceInfo)

			if si.Version == val.Version && si.hasSameTags(val.Tags) {
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
	si, err := sd.decodeServiceInfo(*r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	infos := sd.findExactByServiceInfo(*si)
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
	si, err := sd.decodeServiceInfo(*r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	sd.deleteByServiceInfo(*si)

	w.WriteHeader(http.StatusOK)
}

func (sd *ServiceDiscovery) listServicesHandler(w http.ResponseWriter, r *http.Request) {
	si, err := sd.decodeServiceInfo(*r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}
	infos := sd.findByServiceInfo(*si)
	data, err := json.Marshal(infos)
	if err != nil {
		logger.Log().ErrorWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot convert service info array into JSON")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
