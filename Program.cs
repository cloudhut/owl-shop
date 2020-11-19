using Confluent.Kafka;
using owl_shop.Entities.Order;
using owl_shop.Producers;
using owl_shop.YamlDotNet;
using System;
using System.Collections.Generic;
using System.IO;
using System.Reflection.Metadata;
using System.Threading;
using System.Threading.Tasks;
using YamlDotNet.RepresentationModel;
using YamlDotNet.Serialization;
using YamlDotNet.Serialization.NamingConventions;
using YamlDotNet.Serialization.NodeDeserializers;

namespace owl_shop
{
	class Program
	{
		static OwlJsonSerializer _json = new OwlJsonSerializer();
		static StringSerializer _string = new StringSerializer();

		static async Task Main()
		{
			var yamlConfig = File.ReadAllText("");
			var deserializer = new DeserializerBuilder()
				.WithNamingConvention(UnderscoredNamingConvention.Instance)
				.WithNodeDeserializer(inner => new ValidatingNodeDeserializer(inner), s => s.InsteadOf<ObjectNodeDeserializer>())
				.Build();
			var options = deserializer.Deserialize<Options>(yamlConfig);

			await Test(options);
			Console.WriteLine("");
		}

		static async Task Test(Options options)
		{
			var spammer = new KafkaSpammer(
				new ProducerConfig {
					BootstrapServers = options.kafka.BootstrapServers,
					SaslMechanism = SaslMechanism.Plain,
					SaslUsername = options.kafka.Sasl.Username,
					SaslPassword = options.kafka.Sasl.Password,
					SecurityProtocol = SecurityProtocol.SaslSsl,
					ClientId = "OwlShop",
					SslCaLocation = "",
					Debug = "security"
				},
				_string, _json
			);


			for (int i = 0; i < 100; i++) {
				var r = await spammer.Produce<Order>(
					"orders-json",
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
}
