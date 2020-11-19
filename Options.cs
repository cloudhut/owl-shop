using Confluent.Kafka;
using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using YamlDotNet.Serialization;

namespace owl_shop
{
	class Options
	{
		public KafkaOptions kafka { get; set; }
	}

	class KafkaOptions
	{
		public string ClientId { get; set; } = "OwlShop";
		public KafkaSaslOptions Sasl { get; set; }
		public KafkaTLSOptions? Tls { get; set; }
		public string? BootstrapServers { get; set; }
	}

	class KafkaSaslOptions
	{
		[Required()]
		public SaslMechanism Mechanism { get; set; }

		[Required()]
		public string? Username { get; set; }
		public string? Password { get; set; }
	}

	class KafkaTLSOptions
	{
		public bool Enabled { get; set; } = false;
		public string? CaFilepath { get; set; }
	}
}
