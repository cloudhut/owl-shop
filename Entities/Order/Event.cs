using System;
using System.Collections.Generic;
using System.Text;

namespace owl_shop.Entities.Order
{
	class Event
	{
		public EventType EventType { get; set; }
		public Order Order { get; set; }
	}

	enum EventType
	{
		Created,
		OnHold,
		Updated,
		Cancelled,
		PickingComplete,
		Shippable,
		Delivered,
		Failed,
		Completed,
	}
}
