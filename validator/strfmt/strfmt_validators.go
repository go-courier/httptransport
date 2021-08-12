package strfmt

import (
	github_com_go_courier_validator "github.com/go-courier/httptransport/validator"
)

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(ASCIIValidator) }

var ASCIIValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringASCII, "ascii")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(AlphaValidator) }

var AlphaValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringAlpha, "alpha")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(AlphaNumericValidator) }

var AlphaNumericValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringAlphaNumeric, "alpha-numeric", "alphaNumeric")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(AlphaUnicodeValidator) }

var AlphaUnicodeValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringAlphaUnicode, "alpha-unicode", "alphaUnicode")

func init() {
	github_com_go_courier_validator.ValidatorMgrDefault.Register(AlphaUnicodeNumericValidator)
}

var AlphaUnicodeNumericValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringAlphaUnicodeNumeric, "alpha-unicode-numeric", "alphaUnicodeNumeric")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Base64Validator) }

var Base64Validator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringBase64, "base64")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Base64URLValidator) }

var Base64URLValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringBase64URL, "base64-url", "base64URL")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(BtcAddressValidator) }

var BtcAddressValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringBtcAddress, "btc-address", "btcAddress")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(BtcAddressLowerValidator) }

var BtcAddressLowerValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringBtcAddressLower, "btc-address-lower", "btcAddressLower")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(BtcAddressUpperValidator) }

var BtcAddressUpperValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringBtcAddressUpper, "btc-address-upper", "btcAddressUpper")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(DataURIValidator) }

var DataURIValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringDataURI, "data-uri", "dataURI")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(EmailValidator) }

var EmailValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringEmail, "email")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(EthAddressValidator) }

var EthAddressValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringEthAddress, "eth-address", "ethAddress")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(EthAddressLowerValidator) }

var EthAddressLowerValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringEthAddressLower, "eth-address-lower", "ethAddressLower")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(EthAddressUpperValidator) }

var EthAddressUpperValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringEthAddressUpper, "eth-address-upper", "ethAddressUpper")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(HslValidator) }

var HslValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringHSL, "hsl")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(HslaValidator) }

var HslaValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringHSLA, "hsla")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(HexAdecimalValidator) }

var HexAdecimalValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringHexAdecimal, "hex-adecimal", "hexAdecimal")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(HexColorValidator) }

var HexColorValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringHexColor, "hex-color", "hexColor")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(HostnameValidator) }

var HostnameValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringHostname, "hostname")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(HostnameXValidator) }

var HostnameXValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringHostnameX, "hostname-x", "hostnameX")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Isbn10Validator) }

var Isbn10Validator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringISBN10, "isbn10")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Isbn13Validator) }

var Isbn13Validator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringISBN13, "isbn13")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(LatitudeValidator) }

var LatitudeValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringLatitude, "latitude")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(LongitudeValidator) }

var LongitudeValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringLongitude, "longitude")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(MultibyteValidator) }

var MultibyteValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringMultibyte, "multibyte")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(NumberValidator) }

var NumberValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringNumber, "number")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(NumericValidator) }

var NumericValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringNumeric, "numeric")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(PrintableASCIIValidator) }

var PrintableASCIIValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringPrintableASCII, "printable-ascii", "printableASCII")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(RgbValidator) }

var RgbValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringRGB, "rgb")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(RgbaValidator) }

var RgbaValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringRGBA, "rgba")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(SsnValidator) }

var SsnValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringSSN, "ssn")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(UUIDValidator) }

var UUIDValidator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringUUID, "uuid")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Uuid3Validator) }

var Uuid3Validator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringUUID3, "uuid3")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Uuid4Validator) }

var Uuid4Validator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringUUID4, "uuid4")

func init() { github_com_go_courier_validator.ValidatorMgrDefault.Register(Uuid5Validator) }

var Uuid5Validator = github_com_go_courier_validator.NewRegexpStrfmtValidator(regexpStringUUID5, "uuid5")
