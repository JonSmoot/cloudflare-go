package cloudflare
import (
	"fmt"
	"encoding/json"
	"net/http"
	"io"
	"strings"
	"net/url"
	"crypto/tls"
)

const ContentTypeApplicationJson    string = "application/json"
const ApiUrlBase                    string = "https://api.cloudflare.com/client/v4"
const APiUrlZones                   string = "/zones/"
const ApiUrlDnsRecords              string = "/dns_records/"

// Base Client config, email and api key are mandatory
type Config struct {
	email              string
	key                string
	contentType        string
	insecureSkipVerify bool
}
// Base Service, contains config and defines common methods
type BaseSvc		struct {
	config 			*Config
}
// Zones Service
type ZonesSvc       struct {
	BaseSvc
}
// DNSRecord Service
type DNSRecordsSvc  struct {
	BaseSvc
}

// Core Type decelerations

// Zone type
type Zone struct {
	Name                string
	Id                  string
	Type                string
	NameServers         []string `json:"name_servers"`
	Status              string
	Paused              bool
	OriginalNameServers []string `json:"original_name_servers"`
	DevelopmentMode     uint64   `json:"development_mode"`
}

// DNSRecord type
type DNSRecord struct {
	Id         string    `json:"id,omitempty"`
	Type       string    `json:"type,omitempty"`
	Name       string    `json:"name,omitempty"`
	Content    string    `json:"content,omitempty"`
	Proxiable  bool    	 `json:"proxiable,omitempty"`
	Proxied    bool    	 `json:"proxied,omitempty"`
	Ttl        uint64    `json:"ttl,omitempty"`
	Locked     bool      `json:"locked,omitempty"`
	ZoneId     string    `json:"zone_id,omitempty"`
	ZoneName   string    `json:"zone_name,omitempty"`
	ModifiedOn string    `json:"modified_on,omitempty"`
	CreatedOn  string    `json:"created_on,omitempty"`
	Meta       struct {
				   AutoAdded bool `json:"auto_added,omitempty"`
			   } `json:"meta,omitempty"`
}

// Message type, represents the returned json response from any API invocation
// specific data structures are being decoded from 'Result json.RawMessage' field
type Message struct {
	Success    bool
	Errors     json.RawMessage
	Messages   json.RawMessage
	Result     json.RawMessage
	ResultInfo struct {
				   Page       uint64
				   PerPage    uint64 `json:"per_page"`
				   TotalPages uint64 `json:"total_pages"`
				   Count      uint64
				   TotalCount uint64 `json:"total_count"`
			   } `json:"result_info"`
}

// Helper method for Object to Json String
func JsonString(this interface{}) string {
	out, err := json.Marshal(this)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprint(string(out))
}

// Stringer Implementation for DNSRecord
func (d DNSRecord) String() string {
	return JsonString(d)
}

// Stringer Implementation for Zone
func (z Zone) String() string {
	return JsonString(z)
}

// Config Factory, defaults to application/json type
func NewConfig(email string, key string, contentType string, insecureSkipVerify bool) *Config {
	if email == "" || key == "" {
		return nil
	}
	if contentType == "" {
		contentType = ContentTypeApplicationJson
	}

	return &Config{email: email, key: key, contentType: contentType, insecureSkipVerify: insecureSkipVerify}
}

// ZonesSvc Factory, pass config to inner BaseSvc object
func (config *Config) GetZonesSvc() *ZonesSvc {
	return &ZonesSvc{BaseSvc: BaseSvc{config: config}}
}
// DNSRecordsSvc Factory, pass config to inner BaseSvc object
func (config *Config) GetDNSRecordsSvc() *DNSRecordsSvc {
	return &DNSRecordsSvc{BaseSvc: BaseSvc{config: config}}
}

// Utility function for executing the required http command
func (baseSvc *BaseSvc) Invoke(method string, urlStr string, body io.Reader) (response *http.Response, error error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: baseSvc.config.insecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Auth-Email", baseSvc.config.email)
	req.Header.Add("X-Auth-Key", baseSvc.config.key)
	req.Header.Add("Content-Type", baseSvc.config.contentType)

	return client.Do(req)
}

// Decoder function, Un-marshal results based on received result type
//  Usage:
//		var zones []Zone
//		Decode(resp, &zones)
func (baseSvc *BaseSvc) Decode(resp *http.Response, result interface{}) error {
	var msg Message
	dec := json.NewDecoder(resp.Body)
	err := dec.Decode(&msg)
	if msg.Success {
		err = json.Unmarshal(msg.Result, result)
	}
	return err
}

// Get all zones
//	Usage:
//		config := cloudflare.NewConfig(email, key, "", false)
//		zonesSvc := config.GetZonesSvc()
//		zones, err := zonesSvc.Get()
func (zonesSvc *ZonesSvc) Get() (zones []Zone, err error) {

	resp, err := zonesSvc.Invoke("GET", ApiUrlBase + APiUrlZones, nil)
	if resp == nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = zonesSvc.Decode(resp, &zones)
	return zones, err
}

// Get all DNSRecords for zone
//	Usage:
//		dnsSvc := config.GetDNSRecordsSvc()
//		dnsRecords, err := dnsSvc.Get(zone.Id)
func (dnsSvc DNSRecordsSvc) Get(zoneId string) (dnsRecords []DNSRecord, err error) {
	resp, err := dnsSvc.Invoke("GET", ApiUrlBase + APiUrlZones + zoneId + ApiUrlDnsRecords, nil)
	if resp == nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = dnsSvc.Decode(resp, &dnsRecords)
	return dnsRecords, err
}

// Search DNSRecords for zone
//	Usage:
//		dnsSvc := config.GetDNSRecordsSvc()
//		dnsRecords, err := dnsSvc.Search(zone.Id, "A", "tst.me.com", "", 0, 0)
func (dnsSvc DNSRecordsSvc) Search(zoneId string, typeStr string, name string, content string, page int, perPage int) (dnsRecords []DNSRecord, err error) {
	v := url.Values{}
	if typeStr != "" {
		v.Add("type", typeStr)
	}
	if name != "" {
		v.Add("name", name)
	}
	if content != "" {
		v.Add("content", content)
	}
	if page > 0 {
		v.Add("page", fmt.Sprintf("%d", page))
	}
	if perPage >= 5 && perPage < 100 {
		v.Add("per_page", fmt.Sprintf("%d", perPage))
	}

	resp, err := dnsSvc.Invoke("GET", ApiUrlBase + APiUrlZones + zoneId + ApiUrlDnsRecords + "?" + v.Encode(), nil)
	if resp == nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = dnsSvc.Decode(resp, &dnsRecords)
	return dnsRecords, err
}

// Create new DNSRecords for zone
//	Usage:
//		dnsSvc := config.GetDNSRecordsSvc()
//		dnsRecord := &cloudflare.DNSRecord{Type: "A", Name: "tst.me.com", Content: "98.76.54.321", Ttl: 120, Proxiable: true, Proxied: false}
//		res, err := dnsSvc.Create(zones[0].Id, &dnsRecord)
func (dnsSvc DNSRecordsSvc) Create(zoneId string, dnsRecord *DNSRecord) (res DNSRecord, err error) {
	resp, err := dnsSvc.Invoke("POST", ApiUrlBase + APiUrlZones + zoneId + ApiUrlDnsRecords, strings.NewReader(dnsRecord.String()))
	if resp == nil {
		return DNSRecord{}, err
	}
	defer resp.Body.Close()
	err = dnsSvc.Decode(resp, &res)
	return res, err
}

// Update DNSRecords for zone
//	Usage:
//		dnsSvc := config.GetDNSRecordsSvc()
//		dnsRecord := &cloudflare.DNSRecord{Id: [existing dns record id], Content: "123.45.67.89"}
//		res, err := dnsSvc.Update(zone.Id, &dnsRecord)
func (dnsSvc DNSRecordsSvc) Update(zoneId string, dnsRecord *DNSRecord) (res DNSRecord, err error) {
	resp, err := dnsSvc.Invoke("PUT", ApiUrlBase + APiUrlZones + zoneId + ApiUrlDnsRecords + dnsRecord.Id, strings.NewReader(dnsRecord.String()))
	if resp == nil {
		return DNSRecord{}, err
	}
	defer resp.Body.Close()
	err = dnsSvc.Decode(resp, &res)
	return res, err
}

// Delete DNSRecords for zone
//	Usage:
//		dnsSvc := config.GetDNSRecordsSvc()
//		dnsRecord := &cloudflare.DNSRecord{Id: [existing dns record id]}
//		res, err := dnsSvc.Delete(zone.Id, &dnsRecord)
func (dnsSvc DNSRecordsSvc) Delete(zoneId string, dnsRecord *DNSRecord) (id string, err error) {
	resp, err := dnsSvc.Invoke("DELETE", ApiUrlBase + APiUrlZones + zoneId + ApiUrlDnsRecords + dnsRecord.Id, nil)
	if resp == nil {
		return "", err
	}
	defer resp.Body.Close()
	var retDnsRecord DNSRecord
	err = dnsSvc.Decode(resp, &retDnsRecord)
	return dnsRecord.Id, err
}
