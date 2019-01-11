#!/bin/bash

curl -XPUT -H "Content-Type: application/json" 'http://192.168.3.190:9200/_template/filebeat?pretty' -d@/etc/filebeat/filebeat.template.json
/etc/init.d/filebeat start
tail -f /etc/filebeat/filebeat.template.json