package clob

// CredsCreationWarning is shown when fresh API credentials are minted. They
// cannot be recovered after creation.
const CredsCreationWarning = "Your credentials CANNOT be recovered after they've been created. Be sure to store them safely!"

// InitialCursor / EndCursor are the sentinel cursor values used by the paginated endpoints.
const (
	InitialCursor = "MA=="
	EndCursor     = "LTE="
)

// Bytes32Zero is the canonical zero bytes32 hex string used as default for
// metadata and builder fields on V2 orders.
const Bytes32Zero = "0x0000000000000000000000000000000000000000000000000000000000000000"

// OrderVersionMismatchError is the wire-level error code returned by the CLOB
// when an order is submitted to the wrong-version endpoint.
const OrderVersionMismatchError = "order_version_mismatch"

// BuilderFeesBps is the bps unit used by builder fee rates.
const BuilderFeesBps = 10000
