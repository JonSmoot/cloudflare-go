# cloudflare-go

#### Simple Go Language Client for Cloudflare
> NOTE: This is WIP, currently implemented DNSRecords CRUD operations

## Documentation
See https://godoc.org/github.com/JonSmoot/cloudflare-go

## Usage

> Create Client configuration with your CloudFlare email and key.
> Retrieve the relevant Zone ID and perform DNS Record operations

#### Create Configuration
```go
	import "github.com/JonSmoot/cloudflare-go"
	config := cloudflare.NewConfig(""<<YOUR_EMAIL_ADDRESS>>", "<<YOUR_API_KEY>>","", false)
```

#### Retrieve ZoneService and get all zones
```go
    zonesSvc := config.GetZonesSvc()
	zones, err := zonesSvc.Get()
```

#### Search for DNS Record 
```go
	dnsSvc := config.GetDNSRecordsSvc()
	dnsRecords, err := dnsSvc.Search(zones[0].Id, "A", "ww1.mydomain.com", "", 0, 0)
```

#### Update Content for a DNS Record
```go
    dnsRecords[0].Content = ip
	res, err := dnsSvc.Update(zones[0].Id, &dnsRecords[0])
```


