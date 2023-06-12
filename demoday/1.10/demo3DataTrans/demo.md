# 场景1: 数组数据接收与转换
* stream：
```bash
    bin/kuiper create stream demo '() WITH (datasource="device/gr/messages", format="json")'
```
* rule:

**rule1**
```json
{
  "sql": "SELECT unnest(data), devices from demo;",
  "actions": [{
    "mqtt":  {
      "server": "tcp://broker.emqx.io:1883",
      "topic": "device/gr/results"
    }
  }]
}
```
**rule2**
```json
{
  "sql": "SELECT data, unnest(devices) from demo;",
  "actions": [{
    "mqtt":  {
      "server": "tcp://broker.emqx.io:1883",
      "topic": "device/gr/results"
    }
  }]
}
```
**rule3**
```json
{
  "sql": "SELECT unnest(data), unnest(devices) from demo;",
  "actions": [{
    "mqtt":  {
      "server": "tcp://broker.emqx.io:1883",
      "topic": "device/gr/results"
    }
  }]
}
```

* data:
```json
{
  "data": 
      [{
        "id": "device1",
        "temperature": 20,
        "humidity": 30,
        "time": "2020-06-13 12:00:00",
        "location": "beijing",
        "status": "normal"
      }, 
      {
        "id": "device2",
        "temperature": 22,
        "humidity": 32,
        "time": "2020-06-13 12:00:00",
        "location": "beijing",
        "status": "stop"
      },
      {
        "id": "device3",
        "temperature": 23,
        "humidity": 33,
        "time": "2020-06-13 12:00:00",
        "location": "beijing",
        "status": "normal"
      }],
  "correlationId": "1234567890",
  "timestamp": "2020-06-13 12:00:00",
  "devices": ["device1", "device2", "device3"]
}
```
* results:
  * rule1:
  ```json
    [{"devices":["device1","device2","device3"],"humidity":30,"id":"device1","location":"beijing","status":"normal","temperature":20,"time":"2020-06-13 12:00:00"}]
    [{"devices":["device1","device2","device3"],"humidity":32,"id":"device2","location":"beijing","status":"stop","temperature":22,"time":"2020-06-13 12:00:00"}]
    [{"devices":["device1","device2","device3"],"humidity":33,"id":"device3","location":"beijing","status":"normal","temperature":23,"time":"2020-06-13 12:00:00"}]
  ```
  * rule2:
    ```json
    [{"data":[{"humidity":30,"id":"device1","location":"beijing","status":"normal","temperature":20,"time":"2020-06-13 12:00:00"},{"humidity":32,"id":"device2","location":"beijing","status":"stop","temperature":22,"time":"2020-06-13 12:00:00"},{"humidity":33,"id":"device3","location":"beijing","status":"normal","temperature":23,"time":"2020-06-13 12:00:00"}],"unnest":"device1"}]
    [{"data":[{"humidity":30,"id":"device1","location":"beijing","status":"normal","temperature":20,"time":"2020-06-13 12:00:00"},{"humidity":32,"id":"device2","location":"beijing","status":"stop","temperature":22,"time":"2020-06-13 12:00:00"},{"humidity":33,"id":"device3","location":"beijing","status":"normal","temperature":23,"time":"2020-06-13 12:00:00"}],"unnest":"device2"}]
    [{"data":[{"humidity":30,"id":"device1","location":"beijing","status":"normal","temperature":20,"time":"2020-06-13 12:00:00"},{"humidity":32,"id":"device2","location":"beijing","status":"stop","temperature":22,"time":"2020-06-13 12:00:00"},{"humidity":33,"id":"device3","location":"beijing","status":"normal","temperature":23,"time":"2020-06-13 12:00:00"}],"unnest":"device3"}]
    ```
  * rule3:
    ```json
    {"unnest":"device1"}
    {"unnest":"device2"}
    {"unnest":"device3"}
    多行函数在 SELECT 子句中只允许使用一个多行函数!
    ```
    
# 场景2: 嵌套数据访问与抽取
* stream：
```bash
    bin/kuiper create stream demo '() WITH (datasource="device/gr/messages", format="json")'
```
* rule:

**rule1**
```json
{
  "sql": "SELECT device.data.temperature as temperature, device.data->humidity as humidity, device->data->time as time from demo;",
  "actions": [{
    "mqtt":  {
      "server": "tcp://broker.emqx.io:1883",
      "topic": "device/gr/results"
    }
  }]
}
```

**rule2**
```json
{
  "sql": "SELECT * from demo;",
  "actions": [{
    "mqtt":  {
      "server": "tcp://broker.emqx.io:1883",
      "topic": "device/gr/results",
      "format": "json",
      "dataTemplate": "{{toJson .device}}",
      "dataField": "data",
      "fields": ["temperature", "humidity"],
      "sendSingle": true
    }
  }]
}
```

* data:
```json
{
  "device": {
    "id": 1,
    "data": {
      "temperature": 20,
        "humidity": 30,
      "time": "2020-06-13 12:00:00"
    },
    "location": "beijing"
  },
  "correlationId": "1234567890"
}
```

* results:
  * rule1:
  ```json
    [{"humidity":30,"temperature":20,"time":"2020-06-13 12:00:00"}]
  ```
  * rule2:
  ```json
    {"humidity":30,"temperature":20}
  ```