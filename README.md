# meta-rules-converter

## Convert geo to meta-rule-set

```shell
./converter geosite -f ../geosite.dat -o ../meta-rule/geo/geosite
./converter geoip -f ../geoip.dat -o ../meta-rule/geo/geoip
./converter asn -f ../GeoLite2-ASN.mmdb -o ../meta-rule/asn
```

## Convert geo to sing-rule-set

```shell
./converter geosite -f ../geosite.dat -o ../sing-rule/geo/geosite -t sing-box
./converter geoip -f ../geoip.dat -o ../sing-rule/geo/geoip -t sing-box
./converter asn -f ../GeoLite2-ASN.mmdb -o ../sing-rule/asn -t sing-box
```