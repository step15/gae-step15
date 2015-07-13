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

func ExampleStripLotsOfTags() {
	html := `
<html>
<head>
<title>getword</title>
</head>
<body>
<h1>Your word is</h1>
mysterious
</body>
</html>`
	fmt.Println(StripTags(html))
	// Output: Your word is mysterious
}

func ExampleDropTagsBoring() {
	fmt.Println(dropTags("boring"))
	// Output: boring
}

func ExampleDropTags() {
	fmt.Println(dropTags("please<p>drop<p>me<p>"))
	// Output: please drop me
}
