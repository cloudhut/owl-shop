using Bogus;
using System;
using System.Collections.Generic;
using System.Text;

namespace owl_shop.Entities.Order
{
	class Payment
	{
		public string ID { get; set; }
		public PaymentMethod Method { get; set; }
		public string TransactionID { get; set; }

		public static Payment CreateRandom()
		{
			return new Faker<Payment>()
				.StrictMode(true)
				.RuleFor(p => p.ID, f => f.Random.Guid().ToString())
				.RuleFor(p => p.Method, f => f.Random.Enum<PaymentMethod>())
				.RuleFor(p => p.TransactionID, f => f.Random.Guid().ToString())
				.Generate();
		}
	}

	enum PaymentMethod
	{
		PayPal,
		CreditCard,
		Cash,
		DirectDebit,
	}
}
