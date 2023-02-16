package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"os"
	"sort"
)

const BROKER = "tcp://122.9.166.75:1883"

func loadLabels() ([]string, error) {
	labels := []string{}
	f, err := os.Open("./labels.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	return labels, nil
}

type result struct {
	score float64
	index int
}

func bestMatchLabel(keyValue map[string]interface{}) (string, bool) {
	labels, _ := loadLabels()
	resultArray := keyValue["tfLite"].([]interface{})
	outputString := resultArray[0].(string)

	outputArray, _ := base64.StdEncoding.DecodeString(outputString)

	outputSize := len(outputArray)

	var results []result
	for i := 0; i < outputSize; i++ {
		score := float64(outputArray[i]) / 255.0
		if score < 0.2 {
			continue
		}
		results = append(results, result{score: score, index: i})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	// output is the biggest score labelImage
	if len(results) > 0 {
		return labels[results[0].index], true
	} else {
		return "", true
	}
}

func main() {
	const SUB_TOPIC = "demo1TfliteVideo_result"

	var signal = make(chan bool)

	opts := mqtt.NewClientOptions().AddBroker(BROKER)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	client.Subscribe(SUB_TOPIC, 0, func(client mqtt.Client, message mqtt.Message) {
		resultByte := message.Payload()
		var result map[string]interface{}
		_ = json.Unmarshal(resultByte, &result)
		//
		//fmt.Printf("################ %v\n", result)

		re, _ := bestMatchLabel(result)

		fmt.Printf("################ %v\n", re)

		signal <- true
	})

	<-signal

	client.Disconnect(0)
}
