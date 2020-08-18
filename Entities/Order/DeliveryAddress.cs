using Bogus;
using Bogus.DataSets;
using Bogus.Extensions;
using System;
using System.Collections.Generic;
using System.Text;

namespace owl_shop.Entities.Order
{
	class DeliveryAddress
	{
		public Name.Gender Gender { get; set; }
		public string FirstName { get; set; }
		public string LastName { get; set; }
		public string? Company { get; set; }
		public string Street { get; set; }
		public string BuildingNumber { get; set; }
		public string ZipCode { get; set; }
		public string City { get; set; }
		public string? AddressNotes { get; set; }
		public string Country { get; set; }
		public string PhoneNumber { get; set; }

		public static DeliveryAddress CreateRandom()
		{
			return new Faker<DeliveryAddress>()
				.StrictMode(true)
				.RuleFor(d => d.Gender, f => f.Random.Enum<Name.Gender>())
				.RuleFor(d => d.FirstName, (f, d) => f.Name.FirstName(d.Gender))
				.RuleFor(d => d.LastName, (f, d) => f.Name.LastName(d.Gender))
				.RuleFor(d => d.Company, f => f.Company.CompanyName().OrNull(f, 0.98f))
				.RuleFor(d => d.Street, f => f.Address.StreetName())
				.RuleFor(d => d.BuildingNumber, f => f.Address.BuildingNumber())
				.RuleFor(d => d.ZipCode, f => f.Address.ZipCode())
				.RuleFor(d => d.City, f => f.Address.City())
				.RuleFor(d => d.AddressNotes, f => f.Lorem.Sentence(4, 9).OrNull(f))
				.RuleFor(d => d.Country, f => f.Address.CountryCode())
				.RuleFor(d => d.PhoneNumber, f => f.Person.Phone)
				.Generate();
		}
	}

	public enum Gender
	{
		Male,
		Female
	}
}
