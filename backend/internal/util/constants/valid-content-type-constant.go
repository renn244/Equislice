package constants

var ValidContentType = struct {
	JPEG string
	PNG  string
}{
	JPEG: "image/jpeg",
	PNG:  "image/png",
}

var ValidContentTypeList = []string{
	ValidContentType.JPEG,
	ValidContentType.PNG,
}
