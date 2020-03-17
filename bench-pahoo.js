nodejs/package.json{
  "name": "mbus-test",
  "version": "1.0.0",
  "description": "",
  "main": "script.js",
  "scripts": {
    "test": "echo \"Test SmartLabs\" && exit 1"
  },
  "author": "",
  "license": "MIT",
  "dependencies": {
    "moment": "^2.20.1",
    "mqtt": "^2.15.1",
    "mqtt-connection": "^3.2.0"
  }
}
node_modules
const mqtt = require('mqtt');
var moment = require('moment');

const { HOST, PORT, TOKEN, INTERVAL, PUB_TOPIC } = require('./config.js');

// MQTT Broker URL
const URL = `mqtt://${HOST}:${PORT}`

// PENGHITUNG STATISTIK
let globalCount = 0;
let globalCountLBS = 0;
let count = 0;
let countLBS = 0;

/** Terhubung MQTT Broker
 */
const client  = mqtt.connect(URL, {
    username: TOKEN,
    clientId: 'nodejs'
});

/** Menangani event*/
client.on('connect', function () {
    console.log('Connected to', URL);

    // subscribe topic
    client.subscribe('debug_stream/#')
})

/** Menangani error*/
client.on('disconnect', function () {
    console.log('error', arguments);
})

/** Menangani error*/
client.on('error', function (evt) {
    console.log('error', arguments);
})

/** Menangani koneksi*/
client.on('message', function (topic, message) {
    // message is Buffer
    let data = JSON.parse(message.toString()) || {};

    // menerima info penambahan counter
    count++;

    // Memproses pesan
    if (data.id) {
        let params = data.msg.p;
        let pos = data.msg.pos;

        // Cek pesan
        if (params && pos && 'mnc' in params && 'mcc' in params && 'lac' in params && 'cell_id' in params) {
            // pesan LBS diterima
            countLBS++;
            // publikasikan pesan LBS pada spesifik topik
            client.publish(PUB_TOPIC, JSON.stringify({pos: pos, params: params}))
        }
    }
})

// Catatan statistik (untuk interval ms)
const interval = INTERVAL | 1000;

setInterval(() => {
    // jumlah counter
    globalCount += count;
    globalCountLBS += countLBS;

    // catatan status
    console.log(moment().format('YYYY/MM/DD HH:mm:ss'), (count / (interval / 1000)).toFixed(0), 'msgs/sec, lbs:', (countLBS / (interval / 1000)).toFixed(0), 'msgs/sec');

    // reset counters
    count = 0;
    countLBS = 0;
}, interval);
module.exports = {
  HOST: 'mqtt.smartlabs.com',
  PORT: 1883,
  TOKEN: '*********',
  INTERVAL: 10000,
  PUB_TOPIC: 'LabKimia/gas'
}