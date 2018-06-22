package treeicons

import "github.com/gopherjs/vecty"

func Plus() *vecty.HTML {
	return vecty.Tag(
		"svg",
		vecty.Markup(
			vecty.Attribute("fill", "#000000"),
			vecty.Attribute("height", "24"),
			vecty.Attribute("viewBox", "0 0 24 24"),
			vecty.Attribute("width", "24"),
			vecty.Namespace("http://www.w3.org/2000/svg"),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M0 0h24v24H0z"),
				vecty.Attribute("fill", "none"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M13 7h-2v4H7v2h4v4h2v-4h4v-2h-4V7zm-1-5C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
	)
}

func Minus() *vecty.HTML {
	return vecty.Tag(
		"svg",
		vecty.Markup(
			vecty.Attribute("fill", "#000000"),
			vecty.Attribute("height", "24"),
			vecty.Attribute("viewBox", "0 0 24 24"),
			vecty.Attribute("width", "24"),
			vecty.Namespace("http://www.w3.org/2000/svg"),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M0 0h24v24H0z"),
				vecty.Attribute("fill", "none"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M7 11v2h10v-2H7zm5-9C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
	)
}

func Empty() *vecty.HTML {
	return vecty.Tag(
		"svg",
		vecty.Markup(
			vecty.Attribute("fill", "#000000"),
			vecty.Attribute("height", "24"),
			vecty.Attribute("viewBox", "0 0 24 24"),
			vecty.Attribute("width", "24"),
			vecty.Namespace("http://www.w3.org/2000/svg"),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8z"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M0 0h24v24H0z"),
				vecty.Attribute("fill", "none"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
	)
}

func Unknown() *vecty.HTML {
	return vecty.Tag(
		"svg",
		vecty.Markup(
			vecty.Attribute("fill", "#000000"),
			vecty.Attribute("height", "24"),
			vecty.Attribute("viewBox", "0 0 24 24"),
			vecty.Attribute("width", "24"),
			vecty.Namespace("http://www.w3.org/2000/svg"),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M0 0h24v24H0z"),
				vecty.Attribute("fill", "none"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
		vecty.Tag(
			"path",
			vecty.Markup(
				vecty.Attribute("d", "M10 9c-.55 0-1 .45-1 1s.45 1 1 1 1-.45 1-1-.45-1-1-1zm0 4c-.55 0-1 .45-1 1s.45 1 1 1 1-.45 1-1-.45-1-1-1zM7 9.5c-.28 0-.5.22-.5.5s.22.5.5.5.5-.22.5-.5-.22-.5-.5-.5zm3 7c-.28 0-.5.22-.5.5s.22.5.5.5.5-.22.5-.5-.22-.5-.5-.5zm-3-3c-.28 0-.5.22-.5.5s.22.5.5.5.5-.22.5-.5-.22-.5-.5-.5zm3-6c.28 0 .5-.22.5-.5s-.22-.5-.5-.5-.5.22-.5.5.22.5.5.5zM14 9c-.55 0-1 .45-1 1s.45 1 1 1 1-.45 1-1-.45-1-1-1zm0-1.5c.28 0 .5-.22.5-.5s-.22-.5-.5-.5-.5.22-.5.5.22.5.5.5zm3 6c-.28 0-.5.22-.5.5s.22.5.5.5.5-.22.5-.5-.22-.5-.5-.5zm0-4c-.28 0-.5.22-.5.5s.22.5.5.5.5-.22.5-.5-.22-.5-.5-.5zM12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm2-3.5c-.28 0-.5.22-.5.5s.22.5.5.5.5-.22.5-.5-.22-.5-.5-.5zm0-3.5c-.55 0-1 .45-1 1s.45 1 1 1 1-.45 1-1-.45-1-1-1z"),
				vecty.Namespace("http://www.w3.org/2000/svg"),
			),
		),
	)
}
