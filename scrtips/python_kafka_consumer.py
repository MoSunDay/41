# pip3 install kafka-python

from kafka import KafkaConsumer

consumer = KafkaConsumer('test-41', group_id= 'test', bootstrap_servers= ['localhost:9092'])
for msg in consumer:
    print(msg.value)