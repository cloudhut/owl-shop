using Bogus;
using Bogus.Extensions;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace owl_shop.Entities.Order
{
	class Order
	{
		public string ID { get; set; }
		public DateTime CreatedAt { get; set; }
		public DateTime LastUpdatedAt { get; set; }
		public DateTime? DeliveredAt { get; set; }
		public DateTime? CompletedAt { get; set; }

		public Customer Customer { get; set; }
		public long OrderValue { get; set; }
		public LineItem[] LineItems { get; set; }
		public Payment Payment { get; set; }
		public DeliveryAddress DeliveryAddress { get; set; }

		public static Order CreateRandom()
		{
			// TODO: Add other classes
			return new Faker<Order>()
				.StrictMode(true)
				.RuleFor(o => o.ID, f => f.Random.Guid().ToString())
				.RuleFor(o => o.ID, f => f.Random.Guid().ToString())
				.RuleFor(o => o.CreatedAt, f => f.Date.Recent())
				.RuleFor(o => o.LastUpdatedAt, f => f.Date.Recent())
				.RuleFor(o => o.DeliveredAt, f => f.Date.Recent().OrNull(f, .3f))
				.RuleFor(o => o.CompletedAt, f => f.Date.Recent().OrNull(f, .6f))
				.RuleFor(o => o.Customer, f => Customer.CreateRandom())
				.RuleFor(o => o.LineItems, f => Enumerable.Range(0, f.Random.Int(3, 150)).Select(x => LineItem.CreateRandom()).ToArray())
				.RuleFor(o => o.OrderValue, (f, o) => o.LineItems.Sum(x => x.TotalPrice))
				.RuleFor(o => o.Payment, f => Payment.CreateRandom())
				.RuleFor(o => o.DeliveryAddress, f => DeliveryAddress.CreateRandom())
				.Generate();
		}
	}
}
