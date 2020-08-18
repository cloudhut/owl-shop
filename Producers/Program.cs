using Confluent.Kafka;
using Confluent.SchemaRegistry;
using Confluent.SchemaRegistry.Serdes;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;
using owl_shop.Entities.Order;
using owl_shop.Producers;
using System;
using System.Collections.Generic;
using System.Reflection.Metadata;
using System.Threading.Tasks;

namespace owl_shop
{
	class Program
	{
		static OwlJsonSerializer _json = new OwlJsonSerializer();
		static StringSerializer _string = new StringSerializer();

		static void Main()
		{
			Test().GetAwaiter().GetResult();
			Console.WriteLine("");
		}

		static async Task Test()
		{
			var spammer = new KafkaSpammer(
				new ProducerConfig { BootstrapServers = "localhost:9092" },
				_string, _json
			);

			var r = await spammer.Produce<Order>(
				"order-json",
				o => o.ID,
				o => new (string, object)[]
					{
						("producer_service", "owl-shop-v1"),
						("encoding", "json"),
						("documentation", "none"),
					}
				); 
			Console.WriteLine($"produced to '{r.Topic}' to partition {r.Partition.Value} with offset {r.Offset.Value}");
		}
	}
}
