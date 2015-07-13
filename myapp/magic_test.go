package step

import "fmt"

func ExampleStripTags() {
	html := `
<html>
<head>
<title>getword</title>
</head>
<body>

mysterious
</body>
</html>`
	fmt.Println(StripTags(html))
	// Output: mysterious
}
