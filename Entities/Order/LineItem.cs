using Bogus;
using System;
using System.Collections.Generic;
using System.Text;

namespace owl_shop.Entities.Order
{
	class LineItem
	{
		public string ArticleID { get; set; }
		public string Name { get; set; }
		public int Quantity { get; set; }
		public QuantityUnit QuantityUnit { get; set; }
		public long BasePrice { get; set; }
		public long TotalPrice { get; set; }

		public static LineItem CreateRandom()
		{
			return new Faker<LineItem>()
				.StrictMode(true)
				.RuleFor(l => l.ArticleID, f => f.Random.Guid().ToString())
				.RuleFor(l => l.Name, f => f.Commerce.ProductName())
				.RuleFor(l => l.QuantityUnit, f => f.Random.Enum<QuantityUnit>())
				.RuleFor(l => l.Quantity, f => f.Random.Int(1, 1500)) // TODO: Check quantity unit
				.RuleFor(l => l.BasePrice, f => f.Random.Long(0, 10000))
				.RuleFor(l => l.TotalPrice, (f, l) => l.BasePrice * l.Quantity)
				.Generate();
		}
	}

	enum QuantityUnit
	{
		PIECE,
		GRAM,
		KILOGRAM,
		METRE,
		LITRE,
	}
}
