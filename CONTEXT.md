# Commerce Order Context

This context models a customer order from pricing through stock reservation,
shipment, settlement, and refund.

## Language

**Order**:
A customer's commercial commitment containing the products, quantities, prices,
shipping destination, and lifecycle state.
_Avoid_: Purchase, transaction

**Order Line**:
One product and quantity priced as part of an Order.
_Avoid_: Item

**Customer**:
A person or organization that places Orders.
_Avoid_: Client, account

**Product**:
Something offered for sale with a stable identity, description, price, and
availability status.
_Avoid_: Good, article

**Promotion**:
A time-bounded commercial rule that can change an Order's price or add a gift.
_Avoid_: Deal, price hack

**Inventory Reservation**:
A temporary claim on stock for an unpaid Order. It reduces Stock Available but
does not reduce Stock On Hand.
_Avoid_: Stock deduction, allocation

**Stock On Hand**:
The physical quantity recorded in a warehouse.
_Avoid_: Inventory balance

**Stock Reserved**:
The quantity currently held by live Inventory Reservations.
_Avoid_: Sold stock

**Stock Available**:
The quantity that can still be offered for sale: Stock On Hand minus Stock
Reserved.
_Avoid_: Stock On Hand

**Sub-order**:
The portion of an Order fulfilled by one warehouse, including its deterministic
share of order-level amounts.
_Avoid_: Child order, split

**Payment**:
The money accepted from a Customer for an Order, including the gross amount,
discount, shipping charge, and amount paid.
_Avoid_: Settlement

**Refund**:
A new monetary transaction that returns part or all of a Payment and records
how discounts and shipping were handled.
_Avoid_: Payment edit, reversal

**Shipment**:
The dispatched movement of all or part of an Order from a warehouse toward the
Customer.
_Avoid_: Delivery, Sub-order

**Delivery**:
The completed handoff of a Shipment to the Customer.
_Avoid_: Shipment, dispatch
