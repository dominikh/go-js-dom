// Package dom provides GopherJS and Go bindings for the JavaScript DOM APIs.
//
// This package is an in progress effort of providing idiomatic Go
// bindings for the DOM, wrapping the JavaScript DOM APIs. The API is
// neither complete nor frozen yet, but a great amount of the DOM is
// already useable.
//
// While the package tries to be idiomatic Go, it also tries to stick
// closely to the JavaScript APIs, so that one does not need to learn
// a new set of APIs if one is already familiar with it.
//
// One decision that hasn't been made yet is what parts exactly should
// be part of this package. It is, for example, possible that the
// canvas APIs will live in a separate package. On the other hand,
// types such as StorageEvent (the event that gets fired when the
// HTML5 storage area changes) will be part of this package, simply
// due to how the DOM is structured – even if the actual storage APIs
// might live in a separate package. This might require special care
// to avoid circular dependencies.
//
// The documentation for some of the identifiers is based on the
// MDN Web Docs by Mozilla Contributors (https://developer.mozilla.org/en-US/docs/Web/API),
// licensed under CC-BY-SA 2.5 (https://creativecommons.org/licenses/by-sa/2.5/).
//
//
// Getting started
//
// The usual entry point of using the dom package is by using the
// GetWindow() function which will return a Window, from which you can
// get things such as the current Document.
//
//
// Interfaces
//
// The DOM has a big amount of different element and event types, but
// they all follow three interfaces. All functions that work on or
// return generic elements/events will return one of the three
// interfaces Element, HTMLElement or Event. In these interface values
// there will be concrete implementations, such as
// HTMLParagraphElement or FocusEvent. It's also not unusual that
// values of type Element also implement HTMLElement. In all cases,
// type assertions can be used.
//
// Example:
//     el := dom.GetWindow().Document().QuerySelector(".some-element")
//     htmlEl := el.(dom.HTMLElement)
//     pEl := el.(*dom.HTMLParagraphElement)
//
//
// Live collections
//
// Several functions in the JavaScript DOM return "live"
// collections of elements, that is collections that will be
// automatically updated when elements get removed or added to the
// DOM. Our bindings, however, return static slices of elements that,
// once created, will not automatically reflect updates to the DOM.
// This is primarily done so that slices can actually be used, as
// opposed to a form of iterator, but also because we think that
// magically changing data isn't Go's nature and that snapshots of
// state are a lot easier to reason about.
//
// This does not, however, mean that all objects are snapshots.
// Elements, events and generally objects that aren't slices or maps
// are simple wrappers around JavaScript objects, and as such
// attributes as well as method calls will always return the most
// current data. To reflect this behaviour, these bindings use
// pointers to make the semantics clear. Consider the following
// example:
//
//     d := dom.GetWindow().Document()
//     e1 := d.GetElementByID("my-element")
//     e2 := d.GetElementByID("my-element")
//
//     e1.Class().SetString("some-class")
//     println(e1.Class().String() == e2.Class().String())
//
// The above example will print `true`.
//
//
// DOMTokenList
//
// Some objects in the JS API have two versions of attributes, one
// that returns a string and one that returns a DOMTokenList to ease
// manipulation of string-delimited lists. Some other objects only
// provide DOMTokenList, sometimes DOMSettableTokenList. To simplify
// these bindings, only the DOMTokenList variant will be made
// available, by the type TokenList. In cases where the string
// attribute was the only way to completely replace the value, our
// TokenList will provide Set([]string) and SetString(string) methods,
// which will be able to accomplish the same. Additionally, our
// TokenList will provide methods to convert it to strings and slices.
//
//
// Backwards compatibility
//
// This package has a relatively stable API. However, there will be
// backwards incompatible changes from time to time. This is because
// the package isn't complete yet, as well as because the DOM is a
// moving target, and APIs do change sometimes.
//
// While an attempt is made to reduce changing function signatures to
// a minimum, it can't always be guaranteed. Sometimes mistakes in the
// bindings are found that require changing arguments or return
// values.
//
// Interfaces defined in this package may also change on a
// semi-regular basis, as new methods are added to them. This happens
// because the bindings aren't complete and can never really be, as
// new features are added to the DOM.
//
// If you depend on none of the APIs changing unexpectedly, you're
// advised to vendor this package.
package dom // import "honnef.co/go/js/dom"
