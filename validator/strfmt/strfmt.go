package strfmt

// copy from https://github.com/go-playground/validator/blob/v9/regexes.go
const (
	regexpStringAlpha               = "^[a-zA-Z]+$"
	regexpStringAlphaNumeric        = "^[a-zA-Z0-9]+$"
	regexpStringAlphaUnicode        = "^[\\p{L}]+$"
	regexpStringAlphaUnicodeNumeric = "^[\\p{L}\\p{N}]+$"
	regexpStringNumeric             = "^[-+]?[0-9]+(?:\\.[0-9]+)?$"
	regexpStringNumber              = "^[0-9]+$"
	regexpStringHexAdecimal         = "^[0-9a-fA-F]+$"
	regexpStringHexColor            = "^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"
	regexpStringRGB                 = "^rgb\\(\\s*(?:(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])|(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%)\\s*\\)$"
	regexpStringRGBA                = "^rgba\\(\\s*(?:(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])|(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%\\s*,\\s*(?:0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])%)\\s*,\\s*(?:(?:0.[1-9]*)|[01])\\s*\\)$"
	regexpStringHSL                 = "^hsl\\(\\s*(?:0|[1-9]\\d?|[12]\\d\\d|3[0-5]\\d|360)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*\\)$"
	regexpStringHSLA                = "^hsla\\(\\s*(?:0|[1-9]\\d?|[12]\\d\\d|3[0-5]\\d|360)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0|[1-9]\\d?|100)%)\\s*,\\s*(?:(?:0.[1-9]*)|[01])\\s*\\)$"
	regexpStringEmail               = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:\\(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22)))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	regexpStringBase64              = "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	regexpStringBase64URL           = "^(?:[A-Za-z0-9-_]{4})*(?:[A-Za-z0-9-_]{2}==|[A-Za-z0-9-_]{3}=|[A-Za-z0-9-_]{4})$"
	regexpStringISBN10              = "^(?:[0-9]{9}X|[0-9]{10})$"
	regexpStringISBN13              = "^(?:(?:97(?:8|9))[0-9]{10})$"
	regexpStringUUID3               = "^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$"
	regexpStringUUID4               = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	regexpStringUUID5               = "^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	regexpStringUUID                = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	regexpStringASCII               = "^[\x00-\x7F]*$"
	regexpStringPrintableASCII      = "^[\x20-\x7E]*$"
	regexpStringMultibyte           = "[^\x00-\x7F]"
	regexpStringDataURI             = "^data:.+\\/(.+);base64$"
	regexpStringLatitude            = "^[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)$"
	regexpStringLongitude           = "^[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)$"
	regexpStringSSN                 = `^\d{3}[- ]?\d{2}[- ]?\d{4}$`
	regexpStringHostname            = `^[a-zA-Z][a-zA-Z0-9\-\.]+[a-z-Az0-9]$`    // https://tools.ietf.org/html/rfc952
	regexpStringHostnameX           = `^[a-zA-Z0-9][a-zA-Z0-9\-\.]+[a-z-Az0-9]$` // accepts hostname starting with a digit https://tools.ietf.org/html/rfc1123
	regexpStringBtcAddress          = `^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`        // bitcoin address
	regexpStringBtcAddressUpper     = `^BC1[02-9AC-HJ-NP-Z]{7,76}$`              // bitcoin bech32 address https://en.bitcoin.it/wiki/Bech32
	regexpStringBtcAddressLower     = `^bc1[02-9ac-hj-np-z]{7,76}$`              // bitcoin bech32 address https://en.bitcoin.it/wiki/Bech32
	regexpStringEthAddress          = `^0x[0-9a-fA-F]{40}$`
	regexpStringEthAddressUpper     = `^0x[0-9A-F]{40}$`
	regexpStringEthAddressLower     = `^0x[0-9a-f]{40}$`
)
