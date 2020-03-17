import json
import time

import paho.mqtt.client as mqtt

# Statistik penghitungan
msgs_counter = 0
lbs_counter = 0
dt = time.mktime(datetime.datetime.now().timetuple())
qos = 0

# Statistik interval dump 
interval = 10

# Konfigurasi koneksi MQTT
mqtt_host = "SMARTLABS.COM"
mqtt_port = 1883
mqtt_token = "*********"
mqtt_subscribe_topic = "LabKimia/gas"
mqtt_lbs_topic = "LabKimia/gas"


def on_connect(client, userdata, flags, rc):
    print("Connected to ", mqtt_host)

    # Berlangganan topik
    client.subscribe(mqtt_subscribe_topic)


def on_message(client, userdata, msg):
    global msgs_counter
    global dt
    global lbs_counter

    try:
        data = json.loads(msg.payload.decode('utf-8'))

        msgs_counter += 1
        # Proses pesan
        if 'id' in data:
            # split message to fields
            fields = data['msg']
            pos = fields['pos']
            p = fields['p']

            # Check pos and lbs parameter pada pesan
            if p and pos and 'mcc' in p and 'mnc' in p and 'lac' in p \
                    and 'cell_id' in p:
                lbs_counter += 1
                # send lbs message to specified topic
                lbs_msg = json.dumps(dict(pos=pos, params=p))
                client.publish(mqtt_lbs_topic, bytearray(lbs_msg.encode('utf-8')),
                               qos=qos
                )

    except Exception as e:
        pass

    # Dump and reset penghitungan statistik
    ct = time.mktime(datetime.datetime.now().timetuple())
    if ct - dt > interval:
        print(
            "{}: {} msgs/sec, lbs: {} msgs/sec".format(datetime.datetime.now(),
                                                       msgs_counter / interval,
                                                       lbs_counter / interval)
        )

        dt = ct
        msgs_counter = 0
        lbs_counter = 0

# Membuat MQTT Klien
client = mqtt.Client()
client.on_connect = on_connect
client.on_message = on_message

client.username_pw_set(mqtt_token)
client.connect(mqtt_host, mqtt_port, 60)

# Melakukan penghindaran perulangan akses
client.loop_forever()