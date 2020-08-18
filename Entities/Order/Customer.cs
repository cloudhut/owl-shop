using Bogus;
using Bogus.Extensions;
using System;
using System.Collections.Generic;
using System.Text;

namespace owl_shop.Entities.Order
{
	class Customer
	{
		public string ID { get; set; }
		public string FirstName { get; set; }
		public string LastName { get; set; }
		public string? CompanyName { get; set; }
		public string Email { get; set; }
		public CustomerType Type { get; set; }

		public static Customer CreateRandom()
		{
			return new Faker<Customer>()
				.StrictMode(true)
				.RuleFor(c => c.ID, f => f.Random.Uuid().ToString())
				.RuleFor(c => c.FirstName, f => f.Person.FirstName)
				.RuleFor(c => c.LastName, f => f.Person.LastName)
				.RuleFor(c => c.CompanyName, f => f.Company.CompanyName().OrNull(f, .97f))
				.RuleFor(c => c.Email, (f, c) => f.Internet.Email(c.FirstName, c.LastName))
				.RuleFor(c => c.Type, f => f.Random.Enum<CustomerType>())
				.Generate();
		}
	}

	enum CustomerType
	{
		Personal,
		Business,
		NonProfit,
	}
}
