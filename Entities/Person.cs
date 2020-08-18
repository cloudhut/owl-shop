using Newtonsoft.Json;
using System;
using System.ComponentModel.DataAnnotations;

namespace owl_shop
{
	// todo: create more random objects: Order, UserInteraction(click, add to basked, ...), Customer, Address, Product, DeliveryTracking, DeliveryStatus, ...

	class Person
	{
		[FromSet(SetType.FirstNames)]
		public string FirstName { get; set; }
		public string LastName { get; set; }
		[RangeAttribute(5,80)]
		public int Age { get; set; }

		public static Person CreateRandom()
		{
			return new Person
			{
				FirstName = Data.FirstName,
				LastName = Data.FirstName,
				Age = Data.Age
			};
		}
	}
}
