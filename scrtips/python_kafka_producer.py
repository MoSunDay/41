from kafka import KafkaProducer

producer = KafkaProducer(bootstrap_servers=['localhost:9092'])
future = producer.send('test-41', value= b'my_value')
result = future.get(timeout=10)
print(result)