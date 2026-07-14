package errs

const WRONG_ARGS = "wrong args"
const DIAL_ERR = "errors while trying to connect"
const CONV_ERR = "errors while converting from int to str"
const INVALID_PORT = "invalid port"
const PORT_OCCUPIED = "errors while trying to listen, port may be occupied"
const ACCEPT_ERR = "errors while trying to accept the connection"
const PUB_KEY_ERR = "errors while trying to receive public key"
const SEQ_NUM_TOO_LARGE = "received sequence number, but the difference with ours is too large, dropping"
const SEQ_NUM_NOT_FOUND_CACHE = "expected seq number is bigger than the received one, and a corrisponding key was not found in cache"
const WTF_ERR = "how did we get here"
const B64_ERR = "\rerrors while trying to decode base64"
const DEC_ERR = "\rerrors while trying to decrypt"
