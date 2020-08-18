using System;

namespace owl_shop
{
	[AttributeUsage(AttributeTargets.Property)]
	class Range : Attribute
	{
		public double Min { get; set; }
		public double Max { get; set; }
	}

	[AttributeUsage(AttributeTargets.Property)]
	class FromSet : Attribute
	{
		public SetType SetType { get; set; }
		public FromSet(SetType firstNames)
		{
			SetType = firstNames;
		}
	}
	enum SetType
	{
		FirstNames, // todo find json file or something...
		LastNames,
		Colors,
		Shapes,
		CarNames,
		PopularBrands,
		// ...
	}

	static class RandomGen
	{
		public static T Get<T>()
		{
			// create instance

			// loop over all properties

			// inspect the attribute on the property, generate the right random value
			// set it on the object

			// return constructed object

			return default;
		}
	}
}
