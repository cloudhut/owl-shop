using Bogus;
using Confluent.Kafka;
using Newtonsoft.Json;
using Newtonsoft.Json.Converters;
using Newtonsoft.Json.Serialization;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using System.Text;
using System.Threading.Tasks;

namespace owl_shop.Producers
{
	class KafkaSpammer
	{
		private readonly IKafkaSerializer keySerializer;
		private readonly IKafkaSerializer valueSerializer;
		IProducer<byte[], byte[]> producer;

		public KafkaSpammer(ProducerConfig producerConfig, IKafkaSerializer keySerializer, IKafkaSerializer valueSerializer)
		{
			this.producer = new ProducerBuilder<byte[], byte[]>(producerConfig).Build();
			this.keySerializer = keySerializer;
			this.valueSerializer = valueSerializer;

			// using var schemaRegistry = new CachedSchemaRegistryClient(schemaRegistryConfig);
		}

		public async Task<DeliveryResult<byte[], byte[]>> Produce<T>(string topicName, Func<T, object> getKey, Func<T, (string name, object value)[]>? getHeaders = null)
		{
			var createMethod = typeof(T).GetMethod("CreateRandom", BindingFlags.Static | BindingFlags.Public);
			if (createMethod == null)
				throw new InvalidOperationException($"The type {typeof(T).FullName} you've given must have a static method named 'CreateRandom' (but it doesnt!)");

			var obj = (T)createMethod.Invoke(null, null);

			Headers? headers = null;
			if (getHeaders != null)
			{
				headers = new Headers();
				foreach (var pair in getHeaders(obj))
					headers.Add(pair.name, keySerializer.Serialize(pair.value));
			}


			var message = new Message<byte[], byte[]>
			{
				Headers = headers,
				Key = keySerializer.Serialize(getKey(obj)),
				Value = valueSerializer.Serialize(obj),
				Timestamp = Timestamp.Default
			};

			return await producer.ProduceAsync(topicName, message);
		}
	}

	interface IKafkaSerializer
	{
		byte[] Serialize(object obj);
	}

	class StringSerializer : IKafkaSerializer
	{
		public byte[] Serialize(object obj)
		{
			if (obj == null) throw new ArgumentNullException("obj");
			var str = obj as string;
			if (str == null) throw new InvalidOperationException("the string serializer can only serialize strings, but you have given a: " + obj.GetType().FullName);
			return Encoding.UTF8.GetBytes(str);
		}
	}

	class OwlJsonSerializer : IKafkaSerializer
	{
		static JsonSerializerSettings jsonSettings = new JsonSerializerSettings
		{
			ContractResolver = new DefaultContractResolver { NamingStrategy = new CamelCaseNamingStrategy() },
			Formatting = Formatting.Indented,
			Converters = new List<JsonConverter> { new StringEnumConverter() }
		};

		public byte[] Serialize(object obj)
		{
			var jsonString = JsonConvert.SerializeObject(obj, jsonSettings);
			return Encoding.UTF8.GetBytes(jsonString);
		}
	}
}
