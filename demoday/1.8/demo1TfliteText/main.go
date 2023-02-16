package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const BROKER = "tcp://122.9.166.75:1883"

const (
	START   = "<START>"
	PAD     = "<PAD>"
	UNKNOWN = "<UNKNOWN>"
)

var labels []string

const (
	SENTENCE_LEN = 256
)

func MatchLabel(keyValue map[string]interface{}) {
	labels, _ = loadLabels("labels.txt")

	resultArray := keyValue["tfLite"].([]interface{})
	outputArray := resultArray[0].([]interface{})

	outputSize := len(outputArray)

	type rank struct {
		label string
		poll  float64
	}

	var ranks []rank
	for i := 0; i < outputSize; i++ {
		ranks = append(ranks, rank{
			label: labels[i],
			poll:  outputArray[i].(float64),
		})
	}
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].poll < ranks[j].poll
	})
	// output is the biggest score labelImage
	fmt.Printf("#########result %v\n", ranks)
}

func loadDictionary(fname string) (map[string]int, error) {
	f, err := os.Open("vocab.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dic := make(map[string]int)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		if len(line) < 2 {
			continue
		}
		n, err := strconv.Atoi(line[1])
		if err != nil {
			continue
		}
		dic[line[0]] = n
	}
	return dic, nil
}

func loadLabels(fname string) ([]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var labels []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	return labels, nil
}

func main() {
	dic, err := loadDictionary("vocab.txt")
	if err != nil {
		log.Fatal(err)
	}

	var signal = make(chan bool)

	//model := tflite.NewModelFromFile("text_classification.tflite")
	//if model == nil {
	//	log.Println("cannot load model")
	//	return
	//}
	//defer model.Delete()
	//
	//interpreter := tflite.NewInterpreter(model, nil)
	//defer interpreter.Delete()

	re := regexp.MustCompile(" |\\,|\\.|\\!|\\?|\n")

	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		b, _, err := r.ReadLine()
		if err != nil {
			break
		}
		text := string(b)

		tokens := re.Split(strings.TrimSpace(text), -1)
		index := 0
		tmp := make([]float32, SENTENCE_LEN)
		if n, ok := dic[START]; ok {
			tmp[index] = float32(n)
			index++
		}
		for _, word := range tokens {
			if index >= SENTENCE_LEN {
				break
			}

			if v, ok := dic[word]; ok {
				tmp[index] = float32(v)
			} else {
				tmp[index] = float32(dic[UNKNOWN])
			}
			index++
		}

		for i := index; i < SENTENCE_LEN; i++ {
			tmp[i] = float32(dic[PAD])
		}

		request := map[string]interface{}{}
		request["data"] = tmp
		payload, _ := json.Marshal(request)
		TOPIC := "demo1TfliteText"
		SUB_TOPIC := "demo1TfliteText_result"

		opts := mqtt.NewClientOptions().AddBroker(BROKER)
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		if token := client.Publish(TOPIC, 0, false, payload); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		fmt.Println("Published " + text)

		client.Subscribe(SUB_TOPIC, 0, func(client mqtt.Client, message mqtt.Message) {
			resultByte := message.Payload()
			var result map[string]interface{}
			_ = json.Unmarshal(resultByte, &result)

			MatchLabel(result)

			signal <- true
		})

		<-signal
	}
}
