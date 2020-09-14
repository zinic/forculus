package zmapi

import (
	"strings"

	"github.com/zinic/forculus/zoneminder/constants"
	"golang.org/x/net/html"
)

func FindNodeAttrByKey(node *html.Node, key string) (html.Attribute, bool) {
	for _, attr := range node.Attr {
		if strings.ToLower(attr.Key) == key {
			return attr, true
		}
	}

	return html.Attribute{}, false
}

func ExtractCSRFToken(root *html.Node) (string, bool) {
	depthStack := []*html.Node{
		root,
	}

	for len(depthStack) > 0 {
		nextIdx := len(depthStack) - 1
		nextNode := depthStack[nextIdx]
		depthStack = depthStack[:nextIdx]

		if nextNode.Data == "input" {
			if nameAttr, hasName := FindNodeAttrByKey(nextNode, "name"); hasName {
				if strings.ToLower(nameAttr.Val) == constants.CSRFMagicName {
					if valueAttr, hasValue := FindNodeAttrByKey(nextNode, "value"); hasValue {
						return valueAttr.Val, true
					}
				}
			}
		}

		if nextNode.FirstChild != nil {
			childCursor := nextNode.FirstChild

			for childCursor != nil {
				depthStack = append(depthStack, childCursor)
				childCursor = childCursor.NextSibling
			}
		}
	}

	return "", false
}
