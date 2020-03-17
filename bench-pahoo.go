package main

// Import beberapa package atau library yang dibutuhkan
import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/tidwall/gjson"
)

// Penghitung untuk statistik
var msgsCounter uint32 = 0
var lbsCounter uint32 = 0

// Penghitung interval pembuangan
var dumpInterval time.Duration = time.Second * 10

// Waktu pembuangan terakhir
var lastDumpTime = time.Time{}

// Kredensial koneksi MQTT
var mqttBrokerURI string = "tcp://@mqtt.smartlabs.com:1883"
var mqttToken string = "**********"
var mqttSubscribeTopic = "LabKimia/gas"
var mqttLBSTopic = "LabKimia/gas"

// Penangan pesan
var messageHandler MQTT.MessageHandler = func(client MQTT.Client, message MQTT.Message) {
	// Contoh pesan:
	// {"id":4994688,"msg":{"t":1516446150,"f":1073741827,"tp":"ud","pos":{"y":60.942817688,"x":78.8770904541,"z":69,"s":9,"c":318,"sc":15},"i":1,"lc":0,"p":{"flag1":0,"flag2":0,"flag3":0,"flag4":0,"flag5":0,"flag6":0,"flag7":0,"flag8":0,"field1":27537,"field2":18,"field3":0,"field4":0,"field5":0,"field6":27082,"pwr_ext":27,"pwr_int":4.6,"ver":83,"count":0,"adc1":0,"adc2":0,"gsm":"16(1)"}}}

	// Penambahan pesan yang diterima bertambah
	msgsCounter++

	// Pesan proses
	if gjson.GetBytes(message.Payload(), "id").Exists() {
		// lanjut proses pesan
		msg := gjson.GetBytes(message.Payload(), "msg")
		// get `pos`-object
		pos := msg.Get("pos")
		// get `p`(params)-object
		p := msg.Get("p")

		// Jika posisi dan params ada, coba cari lbs params
		if pos.Exists() && p.Exists() {
			// Jika lbs-params ada - incr. konter dan publikasikan ke topik tertentu
			if p.Get("mnc").Exists() && p.Get("mcc").Exists() && p.Get("lac").Exists() && p.Get("cell_id").Exists() {
				// Bangun pesan lbs
				lbsMessage := map[string]interface{}{
					"pos":    pos.Value(),
					"params": p.Value(),
				}
				lbsMessageString, _ := json.Marshal(lbsMessage)

				// Publikasikan pesan LBS ke topik tertentu
				token := client.Publish(mqttLBSTopic, 0, false, lbsMessageString)
				token.Wait()

				// Penghitung pesan LBS bertambah
				lbsCounter++
			}
		}
	}
}

// Jalankan fungsi setiap `d` (waktu. Durasi)
func doEvery(d time.Duration, f func(time.Time, time.Duration)) {
	for x := range time.Tick(d) {
		f(x, d)
	}
}

// Buang dan atur ulang penghitung
func dumpCounters(t time.Time, d time.Duration) {
	// Dump counters to log
	fmt.Printf("%s: %d msgs/sec, lbs: %d msgs/sec\n", t.Format(time.RFC3339), msgsCounter/uint32(d.Seconds()), lbsCounter/uint32(d.Seconds()))

	// Setel ulang penghitung statistik
	msgsCounter = 0
	lbsCounter = 0
}

// Program Utama
func main() {
	// Tetapkan GOMAXPROCS untuk utas pemrosesan batas
	runtime.GOMAXPROCS(1)

	// Buat struct untuk opsi
	opts := MQTT.NewClientOptions().AddBroker(mqttBrokerURI)
	opts.SetDefaultPublishHandler(messageHandler)
	// Atur mode sesi bersih
	opts.SetCleanSession(true)
	// Tetapkan token sebagai nama pengguna
	opts.SetUsername(mqttToken)

	// Terhubung ke Broker
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	// Berlangganan pada topik
	if token := c.Subscribe(mqttSubscribeTopic, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	// Perbaiki waktu sekarang
	lastDumpTime = time.Now()

	// Jalankan dumpInterval di goroutine
	go doEvery(dumpInterval, dumpCounters)

	// Ini seperti `sementara benar`
	for {
		time.Sleep(1 * time.Second)
	}

	// Tak berlangganan atau Unsubscribe
	if token := c.Unsubscribe(mqttSubscribeTopic); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	// Tak terhubung atau Disconnect
	c.Disconnect(250)
}
