# regatta CoreDNS plugin

## Name
*regatta* - enables serving zone data from [Regatta](https://engineering.jamf.com/regatta/) distributed data store

## Description
The *regatta* plugin is used for serving zone data from Regatta distributed data store

## Syntax
```
regatta [ZONES...] {
  fallthrough [ZONES...]
}
```

* fallthrough If zone matches but no record can be generated, pass request to the next plugin
