// +build js
// +build !go1.14

// Package dom provides Go bindings for the JavaScript DOM APIs.
//
// This package is an in progress effort of providing idiomatic Go
// bindings for the DOM, wrapping the JavaScript DOM APIs. The API is
// neither complete nor frozen yet, but a great amount of the DOM is
// already usable.
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
// current data. To reflect this behavior, these bindings use
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
// This package is currently in development and has an unstable API.
// There may be backwards incompatible changes before it graduates
// from its current alpha status.
//
// There may be additional backwards incompatible changes from time to
// time. This is because the package isn't complete yet, as well as
// because the DOM is a moving target, and APIs do change sometimes.
//
// If you depend on none of the APIs changing unexpectedly, you're
// advised to vendor this package.
package dom // import "honnef.co/go/js/dom/v2"

import (
	"image"
	"image/color"
	"strings"
	"syscall/js"
	"time"
)

// toString returns the string representation of o. If o is nil or
// undefined, the empty string will be returned instead.
func toString(o js.Value) string {
	if o == js.Null() || o == js.Undefined() {
		return ""
	}
	return o.String()
}

func callRecover(o js.Value, fn string, args ...interface{}) (err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr != nil {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	o.Call(fn, args...)
	return nil
}

func elementConstructor(o js.Value) js.Value {
	if n := o.Get("node"); n != js.Undefined() {
		// Support elements wrapped in Polymer's DOM APIs.
		return n.Get("constructor")
	}
	return o.Get("constructor")
}

func arrayToObjects(o js.Value) []js.Value {
	var out []js.Value
	for i := 0; i < o.Length(); i++ {
		out = append(out, o.Index(i))
	}
	return out
}

func nodeListToObjects(o js.Value) []js.Value {
	if o.Get("constructor") == js.Global().Get("Array") {
		// Support Polymer's DOM APIs, which uses Arrays instead of
		// NodeLists
		return arrayToObjects(o)
	}
	var out []js.Value
	length := o.Get("length").Int()
	for i := 0; i < length; i++ {
		out = append(out, o.Call("item", i))
	}
	return out
}

func nodeListToNodes(o js.Value) []Node {
	var out []Node
	for _, obj := range nodeListToObjects(o) {
		out = append(out, wrapNode(obj))
	}
	return out
}

func nodeListToElements(o js.Value) []Element {
	var out []Element
	for _, obj := range nodeListToObjects(o) {
		out = append(out, wrapElement(obj))
	}
	return out
}

func nodeListToHTMLElements(o js.Value) []HTMLElement {
	var out []HTMLElement
	for _, obj := range nodeListToObjects(o) {
		out = append(out, wrapHTMLElement(obj))
	}
	return out
}

func WrapDocument(o js.Value) Document {
	return wrapDocument(o)
}

func WrapDocumentFragment(o js.Value) DocumentFragment {
	return wrapDocumentFragment(o)
}

func WrapNode(o js.Value) Node {
	return wrapNode(o)
}

func WrapElement(o js.Value) Element {
	return wrapElement(o)
}

func WrapHTMLElement(o js.Value) HTMLElement {
	return wrapHTMLElement(o)
}

func wrapDocument(o js.Value) Document {
	switch elementConstructor(o) {
	case js.Global().Get("HTMLDocument"):
		return &htmlDocument{&document{&BasicNode{o}}}
	default:
		return &document{&BasicNode{o}}
	}
}

func wrapDocumentFragment(o js.Value) DocumentFragment {
	switch elementConstructor(o) {
	// TODO: do we have any other stuff we want to check
	default:
		return &documentFragment{&BasicNode{o}}
	}
}

func wrapNode(o js.Value) Node {
	if o == js.Null() || o == js.Undefined() {
		return nil
	}
	switch elementConstructor(o) {
	// TODO all the non-element cases
	case js.Global().Get("Text"):
		return Text{&BasicNode{o}}
	default:
		return wrapElement(o)
	}
}

func wrapElement(o js.Value) Element {
	if o == js.Null() || o == js.Undefined() {
		return nil
	}
	switch elementConstructor(o) {
	// TODO all the non-HTML cases
	default:
		return wrapHTMLElement(o)
	}
}

func wrapHTMLElement(o js.Value) HTMLElement {
	if o == js.Null() || o == js.Undefined() {
		return nil
	}
	el := &BasicHTMLElement{&BasicElement{&BasicNode{o}}}
	c := elementConstructor(o)
	switch c {
	case js.Global().Get("HTMLAnchorElement"):
		return &HTMLAnchorElement{BasicHTMLElement: el, URLUtils: &URLUtils{Value: o}}
	case js.Global().Get("HTMLAppletElement"):
		return &HTMLAppletElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLAreaElement"):
		return &HTMLAreaElement{BasicHTMLElement: el, URLUtils: &URLUtils{Value: o}}
	case js.Global().Get("HTMLAudioElement"):
		return &HTMLAudioElement{HTMLMediaElement: &HTMLMediaElement{BasicHTMLElement: el}}
	case js.Global().Get("HTMLBaseElement"):
		return &HTMLBaseElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLBodyElement"):
		return &HTMLBodyElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLBRElement"):
		return &HTMLBRElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLButtonElement"):
		return &HTMLButtonElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLCanvasElement"):
		return &HTMLCanvasElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLDataElement"):
		return &HTMLDataElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLDataListElement"):
		return &HTMLDataListElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLDirectoryElement"):
		return &HTMLDirectoryElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLDivElement"):
		return &HTMLDivElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLDListElement"):
		return &HTMLDListElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLEmbedElement"):
		return &HTMLEmbedElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLFieldSetElement"):
		return &HTMLFieldSetElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLFontElement"):
		return &HTMLFontElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLFormElement"):
		return &HTMLFormElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLFrameElement"):
		return &HTMLFrameElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLFrameSetElement"):
		return &HTMLFrameSetElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLHeadElement"):
		return &HTMLHeadElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLHeadingElement"):
		return &HTMLHeadingElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLHtmlElement"):
		return &HTMLHtmlElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLHRElement"):
		return &HTMLHRElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLIFrameElement"):
		return &HTMLIFrameElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLImageElement"):
		return &HTMLImageElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLInputElement"):
		return &HTMLInputElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLKeygenElement"):
		return &HTMLKeygenElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLLabelElement"):
		return &HTMLLabelElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLLegendElement"):
		return &HTMLLegendElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLLIElement"):
		return &HTMLLIElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLLinkElement"):
		return &HTMLLinkElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLMapElement"):
		return &HTMLMapElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLMediaElement"):
		return &HTMLMediaElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLMenuElement"):
		return &HTMLMenuElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLMetaElement"):
		return &HTMLMetaElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLMeterElement"):
		return &HTMLMeterElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLModElement"):
		return &HTMLModElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLObjectElement"):
		return &HTMLObjectElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLOListElement"):
		return &HTMLOListElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLOptGroupElement"):
		return &HTMLOptGroupElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLOptionElement"):
		return &HTMLOptionElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLOutputElement"):
		return &HTMLOutputElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLParagraphElement"):
		return &HTMLParagraphElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLParamElement"):
		return &HTMLParamElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLPreElement"):
		return &HTMLPreElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLProgressElement"):
		return &HTMLProgressElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLQuoteElement"):
		return &HTMLQuoteElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLScriptElement"):
		return &HTMLScriptElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLSelectElement"):
		return &HTMLSelectElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLSourceElement"):
		return &HTMLSourceElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLSpanElement"):
		return &HTMLSpanElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLStyleElement"):
		return &HTMLStyleElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableElement"):
		return &HTMLTableElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableCaptionElement"):
		return &HTMLTableCaptionElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableCellElement"):
		return &HTMLTableCellElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableDataCellElement"):
		return &HTMLTableDataCellElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableHeaderCellElement"):
		return &HTMLTableHeaderCellElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableColElement"):
		return &HTMLTableColElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableRowElement"):
		return &HTMLTableRowElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTableSectionElement"):
		return &HTMLTableSectionElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTemplateElement"):
		return &HTMLTemplateElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTextAreaElement"):
		return &HTMLTextAreaElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTimeElement"):
		return &HTMLTimeElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTitleElement"):
		return &HTMLTitleElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLTrackElement"):
		return &HTMLTrackElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLUListElement"):
		return &HTMLUListElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLUnknownElement"):
		return &HTMLUnknownElement{BasicHTMLElement: el}
	case js.Global().Get("HTMLVideoElement"):
		return &HTMLVideoElement{HTMLMediaElement: &HTMLMediaElement{BasicHTMLElement: el}}
	case js.Global().Get("HTMLElement"):
		return el
	default:
		return el
	}
}

func getForm(o js.Value) *HTMLFormElement {
	form := wrapHTMLElement(o.Get("form"))
	if form == nil {
		return nil
	}
	return form.(*HTMLFormElement)
}

func getLabels(o js.Value) []*HTMLLabelElement {
	labels := nodeListToElements(o.Get("labels"))
	out := make([]*HTMLLabelElement, len(labels))
	for i, label := range labels {
		out[i] = label.(*HTMLLabelElement)
	}
	return out
}

func getOptions(o js.Value, attr string) []*HTMLOptionElement {
	options := nodeListToElements(o.Get(attr))
	out := make([]*HTMLOptionElement, len(options))
	for i, option := range options {
		out[i] = option.(*HTMLOptionElement)
	}
	return out
}

func GetWindow() Window {
	return &window{js.Global()}
}

type TokenList struct {
	dtl js.Value // the underlying DOMTokenList
	o   js.Value // the object to which the DOMTokenList belongs
	sa  string   // the name of the corresponding string attribute, empty if there isn't one
}

func (tl *TokenList) Length() int { return tl.dtl.Get("length").Int() }

func (tl *TokenList) Item(idx int) string {
	o := tl.dtl.Call("item", idx)
	return toString(o)
}

func (tl *TokenList) Contains(token string) bool {
	return tl.dtl.Call("contains", token).Bool()
}

func (tl *TokenList) Add(token string) {
	tl.dtl.Call("add", token)
}

func (tl *TokenList) Remove(token string) {
	tl.dtl.Call("remove", token)
}

func (tl *TokenList) Toggle(token string) {
	tl.dtl.Call("toggle", token)
}

func (tl *TokenList) String() string {
	if tl.sa != "" {
		return tl.o.Get(tl.sa).String()
	}
	if tl.dtl.Get("constructor") == js.Global().Get("DOMSettableTokenList") {
		return tl.dtl.Get("value").String()
	}
	// We could manually construct the string, but I am not aware of
	// any case where we have neither a string attribute nor
	// DOMSettableTokenList.
	return ""
}

func (tl *TokenList) Slice() []string {
	var out []string
	length := tl.dtl.Get("length").Int()
	for i := 0; i < length; i++ {
		out = append(out, tl.dtl.Call("item", i).String())
	}
	return out
}

// SetString sets the TokenList's value to the space-separated list of
// tokens in s.
func (tl *TokenList) SetString(s string) {
	if tl.sa != "" {
		tl.o.Set(tl.sa, s)
		return
	}
	if tl.dtl.Get("constructor") == js.Global().Get("DOMSettableTokenList") {
		tl.dtl.Set("value", s)
		return
	}
	// This shouldn't be possible
	panic("no way to SetString on this TokenList")
}

// Set sets the TokenList's value to the list of tokens in s.
//
// Individual tokens in s shouldn't countain spaces.
func (tl *TokenList) Set(s []string) {
	tl.SetString(strings.Join(s, " "))
}

type Document interface {
	Node
	ParentNode

	Async() bool
	SetAsync(bool)
	Doctype() DocumentType
	DocumentElement() Element
	DocumentURI() string
	Implementation() DOMImplementation
	LastStyleSheetSet() string
	PreferredStyleSheetSet() string // TODO correct type?
	SelectedStyleSheetSet() string  // TODO correct type?
	StyleSheets() []StyleSheet      // TODO s/StyleSheet/Stylesheet/ ?
	StyleSheetSets() []StyleSheet   // TODO correct type?
	AdoptNode(node Node) Node
	ImportNode(node Node, deep bool) Node
	CreateElement(name string) Element
	CreateElementNS(namespace, name string) Element
	CreateTextNode(s string) *Text
	ElementFromPoint(x, y int) Element
	EnableStyleSheetsForSet(name string)
	GetElementsByClassName(name string) []Element
	GetElementsByTagName(name string) []Element
	GetElementsByTagNameNS(ns, name string) []Element
	GetElementByID(id string) Element
	QuerySelector(sel string) Element
	QuerySelectorAll(sel string) []Element

	CreateDocumentFragment() DocumentFragment
}

type DocumentFragment interface {
	Node
	ParentNode
	QuerySelector(sel string) Element
	QuerySelectorAll(sel string) []Element
	GetElementByID(id string) Element
}

type HTMLDocument interface {
	Document

	ActiveElement() HTMLElement
	Body() HTMLElement
	Cookie() string
	SetCookie(string)
	DefaultView() Window
	DesignMode() bool
	SetDesignMode(bool)
	Domain() string
	SetDomain(string)
	Forms() []*HTMLFormElement
	Head() *HTMLHeadElement
	Images() []*HTMLImageElement
	LastModified() time.Time
	Links() []HTMLElement
	Location() *Location
	Plugins() []*HTMLEmbedElement
	ReadyState() string
	Referrer() string
	Scripts() []*HTMLScriptElement
	Title() string
	SetTitle(string)
	URL() string

	// TODO HTMLDocument methods
}

type documentFragment struct {
	*BasicNode
}

func (d documentFragment) GetElementByID(id string) Element {
	return wrapElement(d.Call("getElementById", id))
}

func (d documentFragment) QuerySelector(sel string) Element {
	return (&BasicElement{&BasicNode{d.Value}}).QuerySelector(sel)
}

func (d documentFragment) QuerySelectorAll(sel string) []Element {
	return (&BasicElement{&BasicNode{d.Value}}).QuerySelectorAll(sel)
}

type document struct {
	*BasicNode
}

type htmlDocument struct {
	*document
}

func (d *htmlDocument) ActiveElement() HTMLElement {
	return wrapHTMLElement(d.Get("activeElement"))
}

func (d *htmlDocument) Body() HTMLElement {
	return wrapHTMLElement(d.Get("body"))
}

func (d *htmlDocument) Cookie() string {
	return d.Get("cookie").String()
}

func (d *htmlDocument) SetCookie(s string) {
	d.Set("cookie", s)
}

func (d *htmlDocument) DefaultView() Window {
	return &window{d.Get("defaultView")}
}

func (d *htmlDocument) DesignMode() bool {
	s := d.Get("designMode").String()
	return s != "off"
}

func (d *htmlDocument) SetDesignMode(b bool) {
	s := "off"
	if b {
		s = "on"
	}
	d.Set("designMode", s)
}

func (d *htmlDocument) Domain() string {
	return d.Get("domain").String()
}

func (d *htmlDocument) SetDomain(s string) {
	d.Set("domain", s)
}

func (d *htmlDocument) Forms() []*HTMLFormElement {
	var els []*HTMLFormElement
	forms := d.Get("forms")
	length := forms.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(forms.Call("item", i)).(*HTMLFormElement))
	}
	return els
}

func (d *htmlDocument) Head() *HTMLHeadElement {
	head := wrapElement(d.Get("head"))
	if head == nil {
		return nil
	}
	return head.(*HTMLHeadElement)
}

func (d *htmlDocument) Images() []*HTMLImageElement {
	var els []*HTMLImageElement
	images := d.Get("images")
	length := images.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(images.Call("item", i)).(*HTMLImageElement))
	}
	return els
}

func (d *htmlDocument) LastModified() time.Time {
	return time.Unix(0, int64(d.Get("lastModified").Call("getTime").Int())*1000000)
}

func (d *htmlDocument) Links() []HTMLElement {
	var els []HTMLElement
	links := d.Get("links")
	length := links.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(links.Call("item", i)))
	}
	return els
}

func (d *htmlDocument) Location() *Location {
	o := d.Get("location")
	return &Location{Value: o, URLUtils: &URLUtils{Value: o}}
}

func (d *htmlDocument) Plugins() []*HTMLEmbedElement {
	var els []*HTMLEmbedElement
	forms := d.Get("plugins")
	length := forms.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(forms.Call("item", i)).(*HTMLEmbedElement))
	}
	return els
}

func (d *htmlDocument) ReadyState() string {
	return d.Get("readyState").String()
}

func (d *htmlDocument) Referrer() string {
	return d.Get("referrer").String()
}

func (d *htmlDocument) Scripts() []*HTMLScriptElement {
	var els []*HTMLScriptElement
	forms := d.Get("scripts")
	length := forms.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(forms.Call("item", i)).(*HTMLScriptElement))
	}
	return els
}

func (d *htmlDocument) Title() string {
	return d.Get("title").String()
}

func (d *htmlDocument) SetTitle(s string) {
	d.Set("title", s)
}

func (d *htmlDocument) URL() string {
	return d.Get("URL").String()
}

func (d document) Async() bool {
	return d.Get("async").Bool()
}

func (d document) SetAsync(b bool) {
	d.Set("async", b)
}

func (d document) Doctype() DocumentType {
	// FIXME implement
	panic("not implemented")
}

func (d document) DocumentElement() Element {
	return wrapElement(d.Get("documentElement"))
}

func (d document) DocumentURI() string {
	return d.Get("documentURI").String()
}

func (d document) Implementation() DOMImplementation {
	// FIXME implement
	panic("not implemented")
}

func (d document) LastStyleSheetSet() string {
	return d.Get("lastStyleSheetSet").String()
}

func (d document) PreferredStyleSheetSet() string {
	return d.Get("preferredStyleSheetSet").String()
}

func (d document) SelectedStyleSheetSet() string {
	return d.Get("selectedStyleSheetSet").String()
}

func (d document) StyleSheets() []StyleSheet {
	// FIXME implement
	panic("not implemented")
}

func (d document) StyleSheetSets() []StyleSheet {
	// FIXME implement
	panic("not implemented")
}

func (d document) AdoptNode(node Node) Node {
	return wrapNode(d.Call("adoptNode", node.Underlying()))
}

func (d document) ImportNode(node Node, deep bool) Node {
	return wrapNode(d.Call("importNode", node.Underlying(), deep))
}

func (d document) CreateDocumentFragment() DocumentFragment {
	return wrapDocumentFragment(d.Call("createDocumentFragment"))
}

func (d document) CreateElement(name string) Element {
	return wrapElement(d.Call("createElement", name))
}

func (d document) CreateElementNS(ns string, name string) Element {
	return wrapElement(d.Call("createElementNS", ns, name))
}

func (d document) CreateTextNode(s string) *Text {
	return wrapNode(d.Call("createTextNode", s)).(*Text)
}

func (d document) ElementFromPoint(x, y int) Element {
	return wrapElement(d.Call("elementFromPoint", x, y))
}

func (d document) EnableStyleSheetsForSet(name string) {
	d.Call("enableStyleSheetsForSet", name)
}

func (d document) GetElementsByClassName(name string) []Element {
	return (&BasicElement{&BasicNode{d.Value}}).GetElementsByClassName(name)
}

func (d document) GetElementsByTagName(name string) []Element {
	return (&BasicElement{&BasicNode{d.Value}}).GetElementsByTagName(name)
}

func (d document) GetElementsByTagNameNS(ns, name string) []Element {
	return (&BasicElement{&BasicNode{d.Value}}).GetElementsByTagNameNS(ns, name)
}

func (d document) GetElementByID(id string) Element {
	return wrapElement(d.Call("getElementById", id))
}

func (d document) QuerySelector(sel string) Element {
	return (&BasicElement{&BasicNode{d.Value}}).QuerySelector(sel)
}

func (d document) QuerySelectorAll(sel string) []Element {
	return (&BasicElement{&BasicNode{d.Value}}).QuerySelectorAll(sel)
}

type URLUtils struct {
	js.Value
}

func (u *URLUtils) Href() string     { return u.Get("href").String() }
func (u *URLUtils) Protocol() string { return u.Get("protocol").String() }
func (u *URLUtils) Host() string     { return u.Get("host").String() }
func (u *URLUtils) Hostname() string { return u.Get("hostname").String() }
func (u *URLUtils) Port() string     { return u.Get("port").String() }
func (u *URLUtils) Pathname() string { return u.Get("pathname").String() }
func (u *URLUtils) Search() string   { return u.Get("search").String() }
func (u *URLUtils) Hash() string     { return u.Get("hash").String() }
func (u *URLUtils) Username() string { return u.Get("username").String() }
func (u *URLUtils) Password() string { return u.Get("password").String() }
func (u *URLUtils) Origin() string   { return u.Get("origin").String() }

func (u *URLUtils) SetHref(v string)     { u.Set("href", v) }
func (u *URLUtils) SetProtocol(v string) { u.Set("protocol", v) }
func (u *URLUtils) SetHost(v string)     { u.Set("host", v) }
func (u *URLUtils) SetHostname(v string) { u.Set("hostname", v) }
func (u *URLUtils) SetPort(v string)     { u.Set("port", v) }
func (u *URLUtils) SetPathname(v string) { u.Set("pathname", v) }
func (u *URLUtils) SetSearch(v string)   { u.Set("search", v) }
func (u *URLUtils) SetHash(v string)     { u.Set("hash", v) }
func (u *URLUtils) SetUsername(v string) { u.Set("username", v) }
func (u *URLUtils) SetPassword(v string) { u.Set("password", v) }

// TODO Location methods

type Location struct {
	js.Value
	*URLUtils
}

type HTMLElement interface {
	Element
	GlobalEventHandlers

	AccessKey() string
	Dataset() map[string]string
	SetAccessKey(string)
	AccessKeyLabel() string
	SetAccessKeyLabel(string)
	ContentEditable() string
	SetContentEditable(string)
	IsContentEditable() bool
	Dir() string
	SetDir(string)
	Draggable() bool
	SetDraggable(bool)
	Lang() string
	SetLang(string)
	OffsetHeight() float64
	OffsetLeft() float64
	OffsetParent() HTMLElement
	OffsetTop() float64
	OffsetWidth() float64
	Style() *CSSStyleDeclaration
	Title() string
	SetTitle(string)
	Blur()
	Click()
	Focus()
}
type SVGElement interface {
	Element
	// TODO
}

type GlobalEventHandlers interface{}

type Window interface {
	EventTarget

	Console() *Console
	Document() Document
	FrameElement() Element
	Location() *Location
	Name() string
	SetName(string)
	InnerHeight() int
	InnerWidth() int
	Length() int
	Opener() Window
	OuterHeight() int
	OuterWidth() int
	ScrollX() int
	ScrollY() int
	Parent() Window
	ScreenX() int
	ScreenY() int
	ScrollMaxX() int
	ScrollMaxY() int
	Top() Window
	History() History
	Navigator() Navigator
	Screen() *Screen
	Alert(string)
	Back()
	Blur()
	CancelAnimationFrame(int)
	ClearInterval(int)
	ClearTimeout(int)
	Close()
	Confirm(string) bool
	Focus()
	Forward()
	GetComputedStyle(el Element, pseudoElt string) *CSSStyleDeclaration
	GetSelection() Selection
	Home()
	MoveBy(dx, dy int)
	MoveTo(x, y int)
	Open(url, name, features string) Window
	OpenDialog(url, name, features string, args []interface{}) Window
	PostMessage(message string, target string, transfer []interface{})
	Print()
	Prompt(prompt string, initial string) string
	RequestAnimationFrame(callback func(time.Duration)) int
	ResizeBy(dw, dh int)
	ResizeTo(w, h int)
	Scroll(x, y int)
	ScrollBy(dx, dy int)
	ScrollByLines(int)
	ScrollTo(x, y int)
	SetCursor(name string)
	SetInterval(fn func(), delay int) int
	SetTimeout(fn func(), delay int) int
	Stop()
	// TODO constructors
}

type window struct {
	// TODO EventTarget
	js.Value
}

func (w *window) Console() *Console {
	return &Console{w.Get("console")}
}

func (w *window) Document() Document {
	return wrapDocument(w.Get("document"))
}

func (w *window) FrameElement() Element {
	return wrapElement(w.Get("frameElement"))
}

func (w *window) Location() *Location {
	o := w.Get("location")
	return &Location{Value: o, URLUtils: &URLUtils{Value: o}}
}

func (w *window) Name() string {
	return w.Get("name").String()
}

func (w *window) SetName(s string) {
	w.Set("name", s)
}

func (w *window) InnerHeight() int {
	return w.Get("innerHeight").Int()
}

func (w *window) InnerWidth() int {
	return w.Get("innerWidth").Int()
}

func (w *window) Length() int {
	return w.Get("length").Int()
}

func (w *window) Opener() Window {
	return &window{w.Get("opener")}
}

func (w *window) OuterHeight() int {
	return w.Get("outerHeight").Int()
}

func (w *window) OuterWidth() int {
	return w.Get("outerWidth").Int()
}

func (w *window) ScrollX() int {
	return w.Get("scrollX").Int()
}

func (w *window) ScrollY() int {
	return w.Get("scrollY").Int()
}

func (w *window) Parent() Window {
	return &window{w.Get("parent")}
}

func (w *window) ScreenX() int {
	return w.Get("screenX").Int()
}

func (w *window) ScreenY() int {
	return w.Get("screenY").Int()
}

func (w *window) ScrollMaxX() int {
	return w.Get("scrollMaxX").Int()
}

func (w *window) ScrollMaxY() int {
	return w.Get("scrollMaxY").Int()
}

func (w *window) Top() Window {
	return &window{w.Get("top")}
}

func (w *window) History() History {
	// FIXME implement
	return nil
}

func (w *window) Navigator() Navigator {
	// FIXME implement
	panic("not implemented")
}

func (w *window) Screen() *Screen {
	return &Screen{Value: w.Get("screen")}
}

func (w *window) Alert(msg string) {
	w.Call("alert", msg)
}

func (w *window) Back() {
	w.Call("back")
}

func (w *window) Blur() {
	w.Call("blur")
}

func (w *window) ClearInterval(id int) {
	w.Call("clearInterval", id)
}

func (w *window) ClearTimeout(id int) {
	w.Call("clearTimeout", id)
}

func (w *window) Close() {
	w.Call("close")
}

func (w *window) Confirm(prompt string) bool {
	return w.Call("confirm", prompt).Bool()
}

func (w *window) Focus() {
	w.Call("focus")
}

func (w *window) Forward() {
	w.Call("forward")
}

// GetComputedStyle returns the values of all CSS properties of an
// element after applying the active stylesheets. pseudoElt specifies
// the pseudo-element to match. For normal elements, it must be the
// empty string.
func (w *window) GetComputedStyle(el Element, pseudoElt string) *CSSStyleDeclaration {
	var optArg interface{}
	if pseudoElt != "" {
		optArg = pseudoElt
	}
	return &CSSStyleDeclaration{w.Call("getComputedStyle", el.Underlying(), optArg)}
}

func (w *window) GetSelection() Selection {
	// FIXME implement
	panic("not implemented")
}

func (w *window) Home() {
	w.Call("home")
}

func (w *window) MoveBy(dx, dy int) {
	w.Call("moveBy", dx, dy)
}

func (w *window) MoveTo(x, y int) {
	w.Call("moveTo", x, y)
}

func (w *window) Open(url, name, features string) Window {
	return &window{w.Call("open", url, name, features)}
}

func (w *window) OpenDialog(url, name, features string, args []interface{}) Window {
	return &window{w.Call("openDialog", url, name, features, args)}
}

func (w *window) PostMessage(message string, target string, transfer []interface{}) {
	w.Call("postMessage", message, target, transfer)
}

func (w *window) Print() {
	w.Call("print")
}

func (w *window) Prompt(prompt string, initial string) string {
	return w.Call("prompt", prompt, initial).String()
}

func (w *window) ResizeBy(dw, dh int) {
	w.Call("resizeBy", dw, dh)
}

func (w *window) ResizeTo(width, height int) {
	w.Call("resizeTo", width, height)
}

func (w *window) Scroll(x, y int) {
	w.Call("scroll", x, y)
}

func (w *window) ScrollBy(dx, dy int) {
	w.Call("scrollBy", dx, dy)
}

func (w *window) ScrollByLines(i int) {
	w.Call("scrollByLines", i)
}

func (w *window) ScrollTo(x, y int) {
	w.Call("scrollTo", x, y)
}

func (w *window) SetCursor(name string) {
	w.Call("setCursor", name)
}

func (w *window) SetInterval(fn func(), delay int) int {
	// TODO(dmitshur): Call wrapper.Release() when the callback
	// gets cancelled via ClearInterval.
	// See TODO comment in window.RequestAnimationFrame for more details.

	wrapper := js.FuncOf(func(js.Value, []js.Value) interface{} {
		fn()
		return nil
	})
	return w.Call("setInterval", wrapper, delay).Int()
}

func (w *window) SetTimeout(fn func(), delay int) int {
	// TODO(dmitshur): Make sure wrapper.Release() gets called
	// even if the callback gets cancelled via ClearTimeout
	// before triggering.
	// See TODO comment in window.RequestAnimationFrame for more details.

	var wrapper js.Func
	wrapper = js.FuncOf(func(js.Value, []js.Value) interface{} {
		fn()
		wrapper.Release()
		return nil
	})
	return w.Call("setTimeout", wrapper, delay).Int()
}

func (w *window) Stop() {
	w.Call("stop")
}

// TODO reuse util.EventTarget

func (w *window) AddEventListener(typ string, useCapture bool, listener func(Event)) js.Func {
	wrapper := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		listener(wrapEvent(args[0]))
		return nil
	})
	w.Call("addEventListener", typ, wrapper, useCapture)
	return wrapper
}

func (w *window) RemoveEventListener(typ string, useCapture bool, listener js.Func) {
	w.Call("removeEventListener", typ, listener, useCapture)
	listener.Release()
}

func (w *window) DispatchEvent(event Event) bool {
	return w.Call("dispatchEvent", event).Bool()
}

func wrapDOMHighResTimeStamp(o js.Value) time.Duration {
	return time.Duration(o.Float() * float64(time.Millisecond))
}

func (w *window) RequestAnimationFrame(callback func(time.Duration)) int {
	// TODO(dmitshur): Make sure wrapper.Release() gets called
	// even if the callback gets cancelled via CancelAnimationFrame
	// before triggering.
	//
	// But first evaluate if the value of doing that is worth
	// the cost of implementing it (it might require a map and
	// a sync.Mutex in window, etc.).
	//
	// An alternative implementation may be to return something
	// more than just an int from RequestAnimationFrame, and
	// and require user to pass that back in to CancelAnimationFrame,
	// so we can avoid the need for a map/mutex. But that requires
	// an API change and it would deviate more from the DOM API.

	var wrapper js.Func
	wrapper = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		callback(wrapDOMHighResTimeStamp(args[0]))
		wrapper.Release()
		return nil
	})
	return w.Call("requestAnimationFrame", wrapper).Int()
}

func (w *window) CancelAnimationFrame(requestID int) {
	w.Call("cancelAnimationFrame", requestID)
}

// TODO all the other window methods

type Selection interface {
	// TODO
}

type Screen struct {
	js.Value
}

func (s *Screen) AvailTop() int    { return s.Get("availTop").Int() }
func (s *Screen) AvailLeft() int   { return s.Get("availLeft").Int() }
func (s *Screen) AvailHeight() int { return s.Get("availHeight").Int() }
func (s *Screen) AvailWidth() int  { return s.Get("availWidth").Int() }
func (s *Screen) ColorDepth() int  { return s.Get("colorDepth").Int() }
func (s *Screen) Height() int      { return s.Get("height").Int() }
func (s *Screen) Left() int        { return s.Get("left").Int() }
func (s *Screen) PixelDepth() int  { return s.Get("pixelDepth").Int() }
func (s *Screen) Top() int         { return s.Get("top").Int() }
func (s *Screen) Width() int       { return s.Get("width").Int() }

type Navigator interface {
	NavigatorID
	NavigatorLanguage
	NavigatorOnLine
	NavigatorGeolocation
	// NavigatorPlugins
	// NetworkInformation
	CookieEnabled() bool
	DoNotTrack() string
	RegisterProtocolHandler(protocol, uri, title string)
}

type NavigatorID interface {
	AppName() string
	AppVersion() string
	Platform() string
	Product() string
	UserAgent() string
}

type NavigatorLanguage interface {
	Language() string
}

type NavigatorOnLine interface {
	Online() bool
}

type NavigatorGeolocation interface {
	Geolocation() Geolocation
}

type Geolocation interface {
	// TODO wrap PositionOptions into something that uses the JS
	// object
	CurrentPosition(success func(Position), err func(PositionError), opts PositionOptions) Position
	WatchPosition(success func(Position), err func(PositionError), opts PositionOptions) int
	ClearWatch(int)
}

type PositionError struct {
	js.Value
}

func (err *PositionError) Code() int { return err.Get("code").Int() }

func (err *PositionError) Error() string {
	return err.Call("message").String()
}

type PositionOptions struct {
	EnableHighAccuracy bool
	Timeout            time.Duration
	MaximumAge         time.Duration
}

type Position struct {
	Coords    *Coordinates
	Timestamp time.Time
}

type Coordinates struct {
	js.Value
}

func (c *Coordinates) Latitude() float64         { return c.Get("latitude").Float() }
func (c *Coordinates) Longitude() float64        { return c.Get("longitude").Float() }
func (c *Coordinates) Altitude() float64         { return c.Get("altitude").Float() }
func (c *Coordinates) Accuracy() float64         { return c.Get("accuracy").Float() }
func (c *Coordinates) AltitudeAccuracy() float64 { return c.Get("altitudeAccuracy").Float() }
func (c *Coordinates) Heading() float64          { return c.Get("heading").Float() }
func (c *Coordinates) Speed() float64            { return c.Get("speed").Float() }

type History interface {
	Length() int
	State() interface{}
	Back()
	Forward()
	Go(offset int)
	PushState(state interface{}, title string, url string)
	ReplaceState(state interface{}, title string, url string)
}

type Console struct {
	js.Value
	// TODO will replace the js/console package
}

type SVGDocument interface{}
type DocumentType interface{}
type DOMImplementation interface{}
type StyleSheet interface{}
type CSSStyleSheet interface{}

type Node interface {
	EventTarget

	Underlying() js.Value
	BaseURI() string
	ChildNodes() []Node
	FirstChild() Node
	LastChild() Node
	NextSibling() Node
	NodeName() string
	NodeType() int
	NodeValue() string
	SetNodeValue(string)
	OwnerDocument() Document
	ParentNode() Node
	ParentElement() Element
	PreviousSibling() Node
	TextContent() string
	SetTextContent(string)
	AppendChild(Node)
	CloneNode(deep bool) Node
	CompareDocumentPosition(Node) int
	Contains(Node) bool
	HasChildNodes() bool
	InsertBefore(which Node, before Node)
	IsDefaultNamespace(string) bool
	IsEqualNode(Node) bool
	LookupPrefix() string
	LookupNamespaceURI(string) string
	Normalize()
	RemoveChild(Node)
	ReplaceChild(newChild, oldChild Node)
}

// Type BasicNode implements the Node interface and is embedded by
// concrete node types and element types.
type BasicNode struct {
	js.Value
}

func (n *BasicNode) Underlying() js.Value {
	return n.Value
}

func (n *BasicNode) AddEventListener(typ string, useCapture bool, listener func(Event)) js.Func {
	wrapper := js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		listener(wrapEvent(args[0]))
		return nil
	})
	n.Call("addEventListener", typ, wrapper, useCapture)
	return wrapper
}

func (n *BasicNode) RemoveEventListener(typ string, useCapture bool, listener js.Func) {
	n.Call("removeEventListener", typ, listener, useCapture)
	listener.Release()
}

func (n *BasicNode) DispatchEvent(event Event) bool {
	return n.Call("dispatchEvent", event).Bool()
}

func (n *BasicNode) BaseURI() string {
	return n.Get("baseURI").String()
}

func (n *BasicNode) ChildNodes() []Node {
	return nodeListToNodes(n.Get("childNodes"))
}

func (n *BasicNode) FirstChild() Node {
	return wrapNode(n.Get("firstChild"))
}

func (n *BasicNode) LastChild() Node {
	return wrapNode(n.Get("lastChild"))
}

func (n *BasicNode) NextSibling() Node {
	return wrapNode(n.Get("nextSibling"))
}

func (n *BasicNode) NodeName() string {
	return n.Get("nodeName").String()
}

func (n *BasicNode) NodeType() int {
	return n.Get("nodeType").Int()
}

func (n *BasicNode) NodeValue() string {
	return toString(n.Get("nodeValue"))
}

func (n *BasicNode) SetNodeValue(s string) {
	n.Set("nodeValue", s)
}

func (n *BasicNode) OwnerDocument() Document {
	// FIXME implement
	panic("not implemented")
}

func (n *BasicNode) ParentNode() Node {
	return wrapNode(n.Get("parentNode"))
}

func (n *BasicNode) ParentElement() Element {
	return wrapElement(n.Get("parentElement"))
}

func (n *BasicNode) PreviousSibling() Node {
	return wrapNode(n.Get("previousSibling"))
}

func (n *BasicNode) TextContent() string {
	return toString(n.Get("textContent"))
}

func (n *BasicNode) SetTextContent(s string) {
	n.Set("textContent", s)
}

func (n *BasicNode) AppendChild(newchild Node) {
	n.Call("appendChild", newchild.Underlying())
}

func (n *BasicNode) CloneNode(deep bool) Node {
	return wrapNode(n.Call("cloneNode", deep))
}

const (
	DocumentPositionDisconnected           = 1
	DocumentPositionPreceding              = 2
	DocumentPositionFollowing              = 4
	DocumentPositionContains               = 8
	DocumentPositionContainedBy            = 16
	DocumentPositionImplementationSpecific = 32
)

func (n *BasicNode) CompareDocumentPosition(other Node) int {
	return n.Call("compareDocumentPosition", other.Underlying()).Int()
}

func (n *BasicNode) Contains(other Node) bool {
	return n.Call("contains", other.Underlying()).Bool()
}

func (n *BasicNode) HasChildNodes() bool {
	return n.Call("hasChildNodes").Bool()
}

func (n *BasicNode) InsertBefore(which Node, before Node) {
	var o interface{}
	if before != nil {
		o = before.Underlying()
	}
	n.Call("insertBefore", which.Underlying(), o)
}

func (n *BasicNode) IsDefaultNamespace(s string) bool {
	return n.Call("isDefaultNamespace", s).Bool()
}

func (n *BasicNode) IsEqualNode(other Node) bool {
	return n.Call("isEqualNode", other.Underlying()).Bool()
}

func (n *BasicNode) LookupPrefix() string {
	return n.Call("lookupPrefix").String()
}

func (n *BasicNode) LookupNamespaceURI(s string) string {
	return toString(n.Call("lookupNamespaceURI", s))
}

func (n *BasicNode) Normalize() {
	n.Call("normalize")
}

func (n *BasicNode) RemoveChild(other Node) {
	n.Call("removeChild", other.Underlying())
}

func (n *BasicNode) ReplaceChild(newChild, oldChild Node) {
	n.Call("replaceChild", newChild.Underlying(), oldChild.Underlying())
}

type Element interface {
	Node
	ParentNode
	ChildNode

	Attributes() map[string]string
	Class() *TokenList
	Closest(string) Element
	ID() string
	SetID(string)
	TagName() string
	GetAttribute(string) string                   // TODO can attributes only be strings?
	GetAttributeNS(ns string, name string) string // can attributes only be strings?
	GetBoundingClientRect() *Rect
	GetElementsByClassName(string) []Element
	GetElementsByTagName(string) []Element
	GetElementsByTagNameNS(ns string, name string) []Element
	HasAttribute(string) bool
	HasAttributeNS(ns string, name string) bool
	Matches(string) bool
	QuerySelector(string) Element
	QuerySelectorAll(string) []Element
	RemoveAttribute(string)
	RemoveAttributeNS(ns string, name string)
	SetAttribute(name string, value string)
	SetAttributeNS(ns string, name string, value string)
	InnerHTML() string
	SetInnerHTML(string)
	OuterHTML() string
	SetOuterHTML(string)
}

// Rect represents a rectangle.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/DOMRect.
type Rect struct {
	js.Value
}

func (r *Rect) X() float64      { return r.Get("x").Float() }
func (r *Rect) Y() float64      { return r.Get("y").Float() }
func (r *Rect) Width() float64  { return r.Get("width").Float() }
func (r *Rect) Height() float64 { return r.Get("height").Float() }
func (r *Rect) Top() float64    { return r.Get("top").Float() }
func (r *Rect) Right() float64  { return r.Get("right").Float() }
func (r *Rect) Bottom() float64 { return r.Get("bottom").Float() }
func (r *Rect) Left() float64   { return r.Get("left").Float() }

func (r *Rect) SetX(v float64)      { r.Set("x", v) }
func (r *Rect) SetY(v float64)      { r.Set("y", v) }
func (r *Rect) SetWidth(v float64)  { r.Set("width", v) }
func (r *Rect) SetHeight(v float64) { r.Set("height", v) }
func (r *Rect) SetTop(v float64)    { r.Set("top", v) }
func (r *Rect) SetRight(v float64)  { r.Set("right", v) }
func (r *Rect) SetBottom(v float64) { r.Set("bottom", v) }
func (r *Rect) SetLeft(v float64)   { r.Set("left", v) }

type ParentNode interface {
	// No properties/methods that aren't experimental
}

type ChildNode interface {
	PreviousElementSibling() Element
	NextElementSibling() Element
}

// Type BasicHTMLElement implements the HTMLElement interface and is
// embedded by concrete HTML element types.
type BasicHTMLElement struct {
	*BasicElement
	// TODO globalEventHandlers
}

func (e *BasicHTMLElement) AccessKey() string {
	return e.Get("accessKey").String()
}

func (e *BasicHTMLElement) Dataset() map[string]string {
	o := e.Get("dataset")
	data := map[string]string{}
	keys := jsKeys(o)
	for _, key := range keys {
		data[key] = o.Get(key).String()
	}
	return data
}

// jsKeys returns the keys of the given JavaScript object.
func jsKeys(o js.Value) []string {
	if o == js.Null() || o == js.Undefined() {
		return nil
	}
	a := js.Global().Get("Object").Call("keys", o)
	s := make([]string, a.Length())
	for i := 0; i < a.Length(); i++ {
		s[i] = a.Index(i).String()
	}
	return s
}

func (e *BasicHTMLElement) SetAccessKey(s string) {
	e.Set("accessKey", s)
}

func (e *BasicHTMLElement) AccessKeyLabel() string {
	return e.Get("accessKeyLabel").String()
}

func (e *BasicHTMLElement) SetAccessKeyLabel(s string) {
	e.Set("accessKeyLabel", s)
}

func (e *BasicHTMLElement) ContentEditable() string {
	return e.Get("contentEditable").String()
}

func (e *BasicHTMLElement) SetContentEditable(s string) {
	e.Set("contentEditable", s)
}

func (e *BasicHTMLElement) IsContentEditable() bool {
	return e.Get("isContentEditable").Bool()
}

func (e *BasicHTMLElement) Dir() string {
	return e.Get("dir").String()
}

func (e *BasicHTMLElement) SetDir(s string) {
	e.Set("dir", s)
}

func (e *BasicHTMLElement) Draggable() bool {
	return e.Get("draggable").Bool()
}

func (e *BasicHTMLElement) SetDraggable(b bool) {
	e.Set("draggable", b)
}

func (e *BasicHTMLElement) Lang() string {
	return e.Get("lang").String()
}

func (e *BasicHTMLElement) SetLang(s string) {
	e.Set("lang", s)
}

func (e *BasicHTMLElement) OffsetHeight() float64 {
	return e.Get("offsetHeight").Float()
}

func (e *BasicHTMLElement) OffsetLeft() float64 {
	return e.Get("offsetLeft").Float()
}

func (e *BasicHTMLElement) OffsetParent() HTMLElement {
	return wrapHTMLElement(e.Get("offsetParent"))
}

func (e *BasicHTMLElement) OffsetTop() float64 {
	return e.Get("offsetTop").Float()
}

func (e *BasicHTMLElement) OffsetWidth() float64 {
	return e.Get("offsetWidth").Float()
}

func (e *BasicHTMLElement) Style() *CSSStyleDeclaration {
	return &CSSStyleDeclaration{e.Get("style")}
}

func (e *BasicHTMLElement) TabIndex() int {
	return e.Get("tabIndex").Int()
}

func (e *BasicHTMLElement) SetTabIndex(i int) {
	e.Set("tabIndex", i)
}

func (e *BasicHTMLElement) Title() string {
	return e.Get("title").String()
}

func (e *BasicHTMLElement) SetTitle(s string) {
	e.Set("title", s)
}

func (e *BasicHTMLElement) Blur() {
	e.Call("blur")
}

func (e *BasicHTMLElement) Click() {
	e.Call("click")
}

func (e *BasicHTMLElement) Focus() {
	e.Call("focus")
}

// Type BasicElement implements the Element interface and is embedded
// by concrete element types and HTML element types.
type BasicElement struct {
	*BasicNode
}

func (e *BasicElement) Attributes() map[string]string {
	o := e.Get("attributes")
	attrs := map[string]string{}
	length := o.Get("length").Int()
	for i := 0; i < length; i++ {
		item := o.Call("item", i)
		attrs[item.Get("name").String()] = item.Get("value").String()
	}
	return attrs
}

func (e *BasicElement) GetBoundingClientRect() *Rect {
	obj := e.Call("getBoundingClientRect")
	return &Rect{Value: obj}
}

func (e *BasicElement) PreviousElementSibling() Element {
	return wrapElement(e.Get("previousElementSibling"))
}

func (e *BasicElement) NextElementSibling() Element {
	return wrapElement(e.Get("nextElementSibling"))
}

func (e *BasicElement) Class() *TokenList {
	return &TokenList{dtl: e.Get("classList"), o: e.Value, sa: "className"}
}

// SetClass sets the element's className attribute to s. Consider
// using the Class method instead.
func (e *BasicElement) SetClass(s string) {
	e.Set("className", s)
}

func (e *BasicElement) Closest(s string) Element {
	return wrapElement(e.Call("closest", s))
}

func (e *BasicElement) ID() string {
	return e.Get("id").String()
}

func (e *BasicElement) SetID(s string) {
	e.Set("id", s)
}

func (e *BasicElement) TagName() string {
	return e.Get("tagName").String()
}

func (e *BasicElement) GetAttribute(name string) string {
	return toString(e.Call("getAttribute", name))
}

func (e *BasicElement) GetAttributeNS(ns string, name string) string {
	return toString(e.Call("getAttributeNS", ns, name))
}

func (e *BasicElement) GetElementsByClassName(s string) []Element {
	return nodeListToElements(e.Call("getElementsByClassName", s))
}

func (e *BasicElement) GetElementsByTagName(s string) []Element {
	return nodeListToElements(e.Call("getElementsByTagName", s))
}

func (e *BasicElement) GetElementsByTagNameNS(ns string, name string) []Element {
	return nodeListToElements(e.Call("getElementsByTagNameNS", ns, name))
}

func (e *BasicElement) HasAttribute(s string) bool {
	return e.Call("hasAttribute", s).Bool()
}

func (e *BasicElement) HasAttributeNS(ns string, name string) bool {
	return e.Call("hasAttributeNS", ns, name).Bool()
}

func (e *BasicElement) Matches(s string) bool {
	return e.Call("matches", s).Bool()
}

func (e *BasicElement) QuerySelector(s string) Element {
	return wrapElement(e.Call("querySelector", s))
}

func (e *BasicElement) QuerySelectorAll(s string) []Element {
	return nodeListToElements(e.Call("querySelectorAll", s))
}

func (e *BasicElement) RemoveAttribute(s string) {
	e.Call("removeAttribute", s)
}

func (e *BasicElement) RemoveAttributeNS(ns string, name string) {
	e.Call("removeAttributeNS", ns, name)
}

func (e *BasicElement) SetAttribute(name string, value string) {
	e.Call("setAttribute", name, value)
}

func (e *BasicElement) SetAttributeNS(ns string, name string, value string) {
	e.Call("setAttributeNS", ns, name, value)
}

func (e *BasicElement) InnerHTML() string {
	return e.Get("innerHTML").String()
}

func (e *BasicElement) SetInnerHTML(s string) {
	e.Set("innerHTML", s)
}

func (e *BasicElement) OuterHTML() string {
	return e.Get("outerHTML").String()
}

func (e *BasicElement) SetOuterHTML(s string) {
	e.Set("outerHTML", s)
}

type HTMLAnchorElement struct {
	*BasicHTMLElement
	*URLUtils
}

func (e *HTMLAnchorElement) HrefLang() string { return e.BasicHTMLElement.Get("hreflang").String() }
func (e *HTMLAnchorElement) Media() string    { return e.BasicHTMLElement.Get("media").String() }
func (e *HTMLAnchorElement) TabIndex() int    { return e.BasicHTMLElement.Get("tabIndex").Int() }
func (e *HTMLAnchorElement) Target() string   { return e.BasicHTMLElement.Get("target").String() }
func (e *HTMLAnchorElement) Text() string     { return e.BasicHTMLElement.Get("text").String() }
func (e *HTMLAnchorElement) Type() string     { return e.BasicHTMLElement.Get("type").String() }
func (e *HTMLAnchorElement) Rel() *TokenList {
	return &TokenList{dtl: e.Get("relList"), o: e.Value, sa: "rel"}
}

func (e *HTMLAnchorElement) SetHrefLang(v string) { e.BasicHTMLElement.Set("hreflang", v) }
func (e *HTMLAnchorElement) SetMedia(v string)    { e.BasicHTMLElement.Set("media", v) }
func (e *HTMLAnchorElement) SetTabIndex(v int)    { e.BasicHTMLElement.Set("tabIndex", v) }
func (e *HTMLAnchorElement) SetTarget(v string)   { e.BasicHTMLElement.Set("target", v) }
func (e *HTMLAnchorElement) SetText(v string)     { e.BasicHTMLElement.Set("text", v) }
func (e *HTMLAnchorElement) SetType(v string)     { e.BasicHTMLElement.Set("type", v) }

type HTMLAppletElement struct {
	*BasicHTMLElement
}

func (e *HTMLAppletElement) Alt() string      { return e.Get("alt").String() }
func (e *HTMLAppletElement) Coords() string   { return e.Get("coords").String() }
func (e *HTMLAppletElement) HrefLang() string { return e.Get("hreflang").String() }
func (e *HTMLAppletElement) Media() string    { return e.Get("media").String() }
func (e *HTMLAppletElement) Search() string   { return e.Get("search").String() }
func (e *HTMLAppletElement) Shape() string    { return e.Get("shape").String() }
func (e *HTMLAppletElement) TabIndex() int    { return e.Get("tabIndex").Int() }
func (e *HTMLAppletElement) Target() string   { return e.Get("target").String() }
func (e *HTMLAppletElement) Type() string     { return e.Get("type").String() }
func (e *HTMLAppletElement) Rel() *TokenList {
	return &TokenList{dtl: e.Get("relList"), o: e.Value, sa: "rel"}
}

func (e *HTMLAppletElement) SetAlt(v string)      { e.Set("alt", v) }
func (e *HTMLAppletElement) SetCoords(v string)   { e.Set("coords", v) }
func (e *HTMLAppletElement) SetHrefLang(v string) { e.Set("hreflang", v) }
func (e *HTMLAppletElement) SetMedia(v string)    { e.Set("media", v) }
func (e *HTMLAppletElement) SetSearch(v string)   { e.Set("search", v) }
func (e *HTMLAppletElement) SetShape(v string)    { e.Set("shape", v) }
func (e *HTMLAppletElement) SetTabIndex(v int)    { e.Set("tabIndex", v) }
func (e *HTMLAppletElement) SetTarget(v string)   { e.Set("target", v) }
func (e *HTMLAppletElement) SetType(v string)     { e.Set("type", v) }

type HTMLAreaElement struct {
	*BasicHTMLElement
	*URLUtils
}

func (e *HTMLAreaElement) Alt() string      { return e.BasicHTMLElement.Get("alt").String() }
func (e *HTMLAreaElement) Coords() string   { return e.BasicHTMLElement.Get("coords").String() }
func (e *HTMLAreaElement) HrefLang() string { return e.BasicHTMLElement.Get("hreflang").String() }
func (e *HTMLAreaElement) Media() string    { return e.BasicHTMLElement.Get("media").String() }
func (e *HTMLAreaElement) Search() string   { return e.BasicHTMLElement.Get("search").String() }
func (e *HTMLAreaElement) Shape() string    { return e.BasicHTMLElement.Get("shape").String() }
func (e *HTMLAreaElement) TabIndex() int    { return e.BasicHTMLElement.Get("tabIndex").Int() }
func (e *HTMLAreaElement) Target() string   { return e.BasicHTMLElement.Get("target").String() }
func (e *HTMLAreaElement) Type() string     { return e.BasicHTMLElement.Get("type").String() }
func (e *HTMLAreaElement) Rel() *TokenList {
	return &TokenList{dtl: e.Get("relList"), o: e.Value, sa: "rel"}
}

func (e *HTMLAreaElement) SetAlt(v string)      { e.Set("alt", v) }
func (e *HTMLAreaElement) SetCoords(v string)   { e.Set("coords", v) }
func (e *HTMLAreaElement) SetHrefLang(v string) { e.Set("hreflang", v) }
func (e *HTMLAreaElement) SetMedia(v string)    { e.Set("media", v) }
func (e *HTMLAreaElement) SetSearch(v string)   { e.Set("search", v) }
func (e *HTMLAreaElement) SetShape(v string)    { e.Set("shape", v) }
func (e *HTMLAreaElement) SetTabIndex(v int)    { e.Set("tabIndex", v) }
func (e *HTMLAreaElement) SetTarget(v string)   { e.Set("target", v) }
func (e *HTMLAreaElement) SetType(v string)     { e.Set("type", v) }

type HTMLAudioElement struct{ *HTMLMediaElement }

type HTMLBRElement struct{ *BasicHTMLElement }

type HTMLBaseElement struct{ *BasicHTMLElement }

func (e *HTMLBaseElement) Href() string {
	return e.Get("href").String()
}

func (e *HTMLBaseElement) Target() string {
	return e.Get("target").String()
}

type HTMLBodyElement struct{ *BasicHTMLElement }

type HTMLButtonElement struct {
	*BasicHTMLElement
}

func (e *HTMLButtonElement) AutoFocus() bool             { return e.Get("autofocus").Bool() }
func (e *HTMLButtonElement) Disabled() bool              { return e.Get("disabled").Bool() }
func (e *HTMLButtonElement) FormAction() string          { return e.Get("formAction").String() }
func (e *HTMLButtonElement) FormEncType() string         { return e.Get("formEncType").String() }
func (e *HTMLButtonElement) FormMethod() string          { return e.Get("formMethod").String() }
func (e *HTMLButtonElement) FormNoValidate() bool        { return e.Get("formNoValidate").Bool() }
func (e *HTMLButtonElement) FormTarget() string          { return e.Get("formTarget").String() }
func (e *HTMLButtonElement) Name() string                { return e.Get("name").String() }
func (e *HTMLButtonElement) TabIndex() int               { return e.Get("tabIndex").Int() }
func (e *HTMLButtonElement) Type() string                { return e.Get("type").String() }
func (e *HTMLButtonElement) ValidationMessage() string   { return e.Get("validationMessage").String() }
func (e *HTMLButtonElement) Value() string               { return e.Get("value").String() }
func (e *HTMLButtonElement) WillValidate() bool          { return e.Get("willValidate").Bool() }
func (e *HTMLButtonElement) Form() *HTMLFormElement      { return getForm(e.Underlying()) }
func (e *HTMLButtonElement) Labels() []*HTMLLabelElement { return getLabels(e.Underlying()) }
func (e *HTMLButtonElement) Validity() *ValidityState    { return &ValidityState{Value: e.Get("validity")} }

func (e *HTMLButtonElement) SetCustomValidity(s string) { e.Call("setCustomValidity", s) }
func (e *HTMLButtonElement) SetAutoFocus(v bool)        { e.Set("autofocus", v) }
func (e *HTMLButtonElement) SetDisabled(v bool)         { e.Set("disabled", v) }
func (e *HTMLButtonElement) SetFormAction(v string)     { e.Set("formAction", v) }
func (e *HTMLButtonElement) SetFormEncType(v string)    { e.Set("formEncType", v) }
func (e *HTMLButtonElement) SetFormMethod(v string)     { e.Set("formMethod", v) }
func (e *HTMLButtonElement) SetFormNoValidate(v bool)   { e.Set("formNoValidate", v) }
func (e *HTMLButtonElement) SetFormTarget(v string)     { e.Set("formTarget", v) }
func (e *HTMLButtonElement) SetName(v string)           { e.Set("name", v) }
func (e *HTMLButtonElement) SetTabIndex(v int)          { e.Set("tabIndex", v) }
func (e *HTMLButtonElement) SetType(v string)           { e.Set("type", v) }
func (e *HTMLButtonElement) SetValue(v string)          { e.Set("value", v) }

func (e *HTMLButtonElement) CheckValidity() bool { return e.Call("checkValidity").Bool() }

type HTMLCanvasElement struct {
	*BasicHTMLElement
}

func (e *HTMLCanvasElement) Width() int  { return e.Get("width").Int() }
func (e *HTMLCanvasElement) Height() int { return e.Get("height").Int() }

func (e *HTMLCanvasElement) SetWidth(v int)  { e.Set("width", v) }
func (e *HTMLCanvasElement) SetHeight(v int) { e.Set("height", v) }

type CanvasRenderingContext2D struct {
	js.Value
}

// Colors, Styles, and Shadows

func (ctx *CanvasRenderingContext2D) FillStyle() string   { return ctx.Get("fillStyle").String() }
func (ctx *CanvasRenderingContext2D) StrokeStyle() string { return ctx.Get("strokeStyle").String() }
func (ctx *CanvasRenderingContext2D) ShadowColor() string { return ctx.Get("shadowColor").String() }
func (ctx *CanvasRenderingContext2D) ShadowBlur() int     { return ctx.Get("shadowBlur").Int() }
func (ctx *CanvasRenderingContext2D) ShadowOffsetX() int  { return ctx.Get("shadowOffsetX").Int() }
func (ctx *CanvasRenderingContext2D) ShadowOffsetY() int  { return ctx.Get("shadowOffsetY").Int() }

func (ctx *CanvasRenderingContext2D) SetFillStyle(v string)   { ctx.Set("fillStyle", v) }
func (ctx *CanvasRenderingContext2D) SetStrokeStyle(v string) { ctx.Set("strokeStyle", v) }
func (ctx *CanvasRenderingContext2D) SetShadowColor(v string) { ctx.Set("shadowColor", v) }
func (ctx *CanvasRenderingContext2D) SetShadowBlur(v int)     { ctx.Set("shadowBlur", v) }
func (ctx *CanvasRenderingContext2D) SetShadowOffsetX(v int)  { ctx.Set("shadowOffsetX", v) }
func (ctx *CanvasRenderingContext2D) SetShadowOffsetY(v int)  { ctx.Set("shadowOffsetY", v) }

// Line Styles

func (ctx *CanvasRenderingContext2D) LineCap() string  { return ctx.Get("lineCap").String() }
func (ctx *CanvasRenderingContext2D) LineJoin() string { return ctx.Get("lineJoin").String() }
func (ctx *CanvasRenderingContext2D) LineWidth() int   { return ctx.Get("lineWidth").Int() }
func (ctx *CanvasRenderingContext2D) MiterLimit() int  { return ctx.Get("miterLimit").Int() }

func (ctx *CanvasRenderingContext2D) SetLineCap(v string)  { ctx.Set("lineCap", v) }
func (ctx *CanvasRenderingContext2D) SetLineJoin(v string) { ctx.Set("lineJoin", v) }
func (ctx *CanvasRenderingContext2D) SetLineWidth(v int)   { ctx.Set("lineWidth", v) }
func (ctx *CanvasRenderingContext2D) SetMiterLimit(v int)  { ctx.Set("miterLimit", v) }

// Text

func (ctx *CanvasRenderingContext2D) Font() string         { return ctx.Get("font").String() }
func (ctx *CanvasRenderingContext2D) TextAlign() string    { return ctx.Get("textAlign").String() }
func (ctx *CanvasRenderingContext2D) TextBaseline() string { return ctx.Get("textBaseline").String() }

func (ctx *CanvasRenderingContext2D) SetFont(v string)         { ctx.Set("font", v) }
func (ctx *CanvasRenderingContext2D) SetTextAlign(v string)    { ctx.Set("textAlign", v) }
func (ctx *CanvasRenderingContext2D) SetTextBaseline(v string) { ctx.Set("textBaseline", v) }

// Compositing

func (ctx *CanvasRenderingContext2D) GlobalAlpha() float64 { return ctx.Get("globalAlpha").Float() }
func (ctx *CanvasRenderingContext2D) GlobalCompositeOperation() string {
	return ctx.Get("globalCompositeOperation").String()
}

func (ctx *CanvasRenderingContext2D) SetGlobalAlpha(v float64) { ctx.Set("globalAlpha", v) }
func (ctx *CanvasRenderingContext2D) SetGlobalCompositeOperation(v string) {
	ctx.Set("globalCompositeOperation", v)
}

type ImageData struct {
	js.Value
}

func (m *ImageData) Width() int     { return m.Get("width").Int() }
func (m *ImageData) Height() int    { return m.Get("height").Int() }
func (m *ImageData) Data() js.Value { return m.Get("data") }

func (m *ImageData) ColorModel() color.Model { return color.NRGBAModel }

func (m *ImageData) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.Get("width").Int(), m.Get("height").Int())
}

func (m *ImageData) At(x, y int) color.Color {
	return m.NRGBAAt(x, y)
}

func (m *ImageData) NRGBAAt(x, y int) color.NRGBA {
	if x < 0 || x >= m.Get("width").Int() ||
		y < 0 || y >= m.Get("height").Int() {
		return color.NRGBA{}
	}
	i := (y*m.Get("width").Int() + x) * 4
	return color.NRGBA{
		R: uint8(m.Get("data").Index(i + 0).Int()),
		G: uint8(m.Get("data").Index(i + 1).Int()),
		B: uint8(m.Get("data").Index(i + 2).Int()),
		A: uint8(m.Get("data").Index(i + 3).Int()),
	}
}

func (m *ImageData) Set(x, y int, c color.Color) {
	if x < 0 || x >= m.Get("width").Int() ||
		y < 0 || y >= m.Get("height").Int() {
		return
	}
	c1 := color.NRGBAModel.Convert(c).(color.NRGBA)
	i := (y*m.Get("width").Int() + x) * 4
	m.Get("data").SetIndex(i+0, c1.R)
	m.Get("data").SetIndex(i+1, c1.G)
	m.Get("data").SetIndex(i+2, c1.B)
	m.Get("data").SetIndex(i+3, c1.A)
}

func (m *ImageData) SetNRGBA(x, y int, c color.NRGBA) {
	if x < 0 || x >= m.Get("width").Int() ||
		y < 0 || y >= m.Get("weight").Int() {
		return
	}
	i := (y*m.Get("width").Int() + x) * 4
	m.Get("data").SetIndex(i+0, c.R)
	m.Get("data").SetIndex(i+1, c.G)
	m.Get("data").SetIndex(i+2, c.B)
	m.Get("data").SetIndex(i+3, c.A)
}

// CanvasGradient represents an opaque object describing a gradient.
// It is returned by the methods CanvasRenderingContext2D.CreateLinearGradient
// or CanvasRenderingContext2D.CreateRadialGradient.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/CanvasGradient.
type CanvasGradient struct {
	js.Value
}

// AddColorStop adds a new stop, defined by an offset and a color, to the gradient.
// It panics with *js.Error if the offset is not between 0 and 1, or if the color
// can't be parsed as a CSS <color>.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/CanvasGradient/addColorStop.
func (cg *CanvasGradient) AddColorStop(offset float64, color string) {
	cg.Call("addColorStop", offset, color)
}

// CanvasPattern represents an opaque object describing a pattern.
// It is based on an image, a canvas or a video, created by the
// CanvasRenderingContext2D.CreatePattern method.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/CanvasPattern.
type CanvasPattern struct {
	js.Value
}

type TextMetrics struct {
	js.Value
}

func (m *TextMetrics) Width() float64                  { return m.Get("width").Float() }
func (m *TextMetrics) ActualBoundingBoxLeft() float64  { return m.Get("actualBoundingBoxLeft").Float() }
func (m *TextMetrics) ActualBoundingBoxRight() float64 { return m.Get("actualBoundingBoxRight").Float() }
func (m *TextMetrics) FontBoundingBoxAscent() float64  { return m.Get("fontBoundingBoxAscent").Float() }
func (m *TextMetrics) FontBoundingBoxDescent() float64 { return m.Get("fontBoundingBoxDescent").Float() }
func (m *TextMetrics) ActualBoundingBoxAscent() float64 {
	return m.Get("actualBoundingBoxAscent").Float()
}
func (m *TextMetrics) ActualBoundingBoxDescent() float64 {
	return m.Get("actualBoundingBoxDescent").Float()
}
func (m *TextMetrics) EmHeightAscent() float64      { return m.Get("emHeightAscent").Float() }
func (m *TextMetrics) EmHeightDescent() float64     { return m.Get("emHeightDescent").Float() }
func (m *TextMetrics) HangingBaseline() float64     { return m.Get("hangingBaseline").Float() }
func (m *TextMetrics) AlphabeticBaseline() float64  { return m.Get("alphabeticBaseline").Float() }
func (m *TextMetrics) IdeographicBaseline() float64 { return m.Get("ideographicBaseline").Float() }

// Creating canvas 2d context

func (e *HTMLCanvasElement) GetContext2d() *CanvasRenderingContext2D {
	ctx := e.GetContext("2d")
	return &CanvasRenderingContext2D{Value: ctx}
}

func (e *HTMLCanvasElement) GetContext(param string) js.Value {
	return e.Call("getContext", param)
}

// Drawing Rectangles

func (ctx *CanvasRenderingContext2D) ClearRect(x, y, width, height float64) {
	ctx.Call("clearRect", x, y, width, height)
}

func (ctx *CanvasRenderingContext2D) FillRect(x, y, width, height float64) {
	ctx.Call("fillRect", x, y, width, height)
}

func (ctx *CanvasRenderingContext2D) StrokeRect(x, y, width, height float64) {
	ctx.Call("strokeRect", x, y, width, height)
}

// Drawing Text

// FillText fills a given text at the given (x, y) position.
// If the optional maxWidth parameter is not -1,
// the text will be scaled to fit that width.
func (ctx *CanvasRenderingContext2D) FillText(text string, x, y, maxWidth float64) {
	if maxWidth == -1 {
		ctx.Call("fillText", text, x, y)
		return
	}

	ctx.Call("fillText", text, x, y, maxWidth)
}

// StrokeText strokes a given text at the given (x, y) position.
// If the optional maxWidth parameter is not -1,
// the text will be scaled to fit that width.
func (ctx *CanvasRenderingContext2D) StrokeText(text string, x, y, maxWidth float64) {
	if maxWidth == -1 {
		ctx.Call("strokeText", text, x, y)
		return
	}

	ctx.Call("strokeText", text, x, y, maxWidth)
}
func (ctx *CanvasRenderingContext2D) MeasureText(text string) *TextMetrics {
	textMetrics := ctx.Call("measureText", text)
	return &TextMetrics{Value: textMetrics}
}

// Line styles

func (ctx *CanvasRenderingContext2D) GetLineDash() []float64 {
	var dashes []float64
	lineDash := ctx.Call("getLineDash")
	for i := 0; i < lineDash.Length(); i++ {
		dashes = append(dashes, lineDash.Index(i).Float())
	}
	return dashes
}

func (ctx *CanvasRenderingContext2D) SetLineDash(dashes []float64) {
	ctx.Call("setLineDash", dashes)
}

// Gradients and patterns

// CreateLinearGradient creates a linear gradient along the line given
// by the coordinates represented by the parameters.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/createLinearGradient.
func (ctx *CanvasRenderingContext2D) CreateLinearGradient(x0, y0, x1, y1 float64) *CanvasGradient {
	return &CanvasGradient{Value: ctx.Call("createLinearGradient", x0, y0, x1, y1)}
}

// CreateRadialGradient creates a radial gradient given by the coordinates of the two circles
// represented by the parameters.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/createRadialGradient.
func (ctx *CanvasRenderingContext2D) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) *CanvasGradient {
	return &CanvasGradient{Value: ctx.Call("createRadialGradient", x0, y0, r0, x1, y1, r1)}
}

// CreatePattern creates a pattern using the specified image (a CanvasImageSource).
// It repeats the source in the directions specified by the repetition argument.
//
// Reference: https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/createPattern.
func (ctx *CanvasRenderingContext2D) CreatePattern(image Element, repetition string) *CanvasPattern {
	return &CanvasPattern{Value: ctx.Call("createPattern", image, repetition)}
}

// Paths

func (ctx *CanvasRenderingContext2D) BeginPath() {
	ctx.Call("beginPath")
}

func (ctx *CanvasRenderingContext2D) ClosePath() {
	ctx.Call("closePath")
}

func (ctx *CanvasRenderingContext2D) MoveTo(x, y float64) {
	ctx.Call("moveTo", x, y)
}

func (ctx *CanvasRenderingContext2D) LineTo(x, y float64) {
	ctx.Call("lineTo", x, y)
}

func (ctx *CanvasRenderingContext2D) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	ctx.Call("bezierCurveTo", cp1x, cp1y, cp2x, cp2y, x, y)
}

func (ctx *CanvasRenderingContext2D) QuadraticCurveTo(cpx, cpy, x, y float64) {
	ctx.Call("quadraticCurveTo", cpx, cpy, x, y)
}

func (ctx *CanvasRenderingContext2D) Arc(x, y, r, sAngle, eAngle float64, counterclockwise bool) {
	ctx.Call("arc", x, y, r, sAngle, eAngle, counterclockwise)
}

func (ctx *CanvasRenderingContext2D) ArcTo(x1, y1, x2, y2, r float64) {
	ctx.Call("arcTo", x1, y1, x2, y2, r)
}

func (ctx *CanvasRenderingContext2D) Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64, anticlockwise bool) {
	ctx.Call("ellipse", x, y, radiusX, radiusY, rotation, startAngle, endAngle, anticlockwise)
}

func (ctx *CanvasRenderingContext2D) Rect(x, y, width, height float64) {
	ctx.Call("rect", x, y, width, height)
}

// Drawing paths

func (ctx *CanvasRenderingContext2D) Fill() {
	ctx.Call("fill")
}

func (ctx *CanvasRenderingContext2D) Stroke() {
	ctx.Call("stroke")
}

func (ctx *CanvasRenderingContext2D) DrawFocusIfNeeded(element HTMLElement, path js.Value) {
	ctx.Call("drawFocusIfNeeded", element, path)
}

func (ctx *CanvasRenderingContext2D) ScrollPathIntoView(path js.Value) {
	ctx.Call("scrollPathIntoView", path)
}

func (ctx *CanvasRenderingContext2D) Clip() {
	ctx.Call("clip")
}

func (ctx *CanvasRenderingContext2D) IsPointInPath(x, y float64) bool {
	return ctx.Call("isPointInPath", x, y).Bool()
}

func (ctx *CanvasRenderingContext2D) IsPointInStroke(path js.Value, x, y float64) bool {
	return ctx.Call("isPointInStroke", path, x, y).Bool()
}

// Transformations

func (ctx *CanvasRenderingContext2D) Rotate(angle float64) {
	ctx.Call("rotate", angle)
}

func (ctx *CanvasRenderingContext2D) Scale(scaleWidth, scaleHeight float64) {
	ctx.Call("scale", scaleWidth, scaleHeight)
}

func (ctx *CanvasRenderingContext2D) Translate(x, y float64) {
	ctx.Call("translate", x, y)
}

func (ctx *CanvasRenderingContext2D) Transform(a, b, c, d, e, f float64) {
	ctx.Call("transform", a, b, c, d, e, f)
}

func (ctx *CanvasRenderingContext2D) SetTransform(a, b, c, d, e, f float64) {
	ctx.Call("setTransform", a, b, c, d, e, f)
}

func (ctx *CanvasRenderingContext2D) ResetTransform() {
	ctx.Call("resetTransform")
}

// Drawing images

func (ctx *CanvasRenderingContext2D) DrawImage(image Element, dx, dy float64) {
	ctx.Call("drawImage", image, dx, dy)
}

func (ctx *CanvasRenderingContext2D) DrawImageWithDst(image Element, dx, dy, dWidth, dHeight float64) {
	ctx.Call("drawImage", image, dx, dy, dWidth, dHeight)
}

func (ctx *CanvasRenderingContext2D) DrawImageWithSrcAndDst(image Element, sx, sy, sWidth, sHeight, dx, dy, dWidth, dHeight float64) {
	ctx.Call("drawImage", image, sx, sy, sWidth, sHeight, dx, dy, dWidth, dHeight)
}

// Pixel manipulation

func (ctx *CanvasRenderingContext2D) CreateImageData(width, height int) *ImageData {
	return &ImageData{Value: ctx.Call("createImageData", width, height)}
}

func (ctx *CanvasRenderingContext2D) GetImageData(sx, sy, sw, sh int) *ImageData {
	return &ImageData{Value: ctx.Call("getImageData", sx, sy, sw, sh)}
}

func (ctx *CanvasRenderingContext2D) PutImageData(imageData *ImageData, dx, dy float64) {
	ctx.Call("putImageData", imageData, dx, dy)
}

func (ctx *CanvasRenderingContext2D) PutImageDataDirty(imageData *ImageData, dx, dy float64, dirtyX, dirtyY, dirtyWidth, dirtyHeight int) {
	ctx.Call("putImageData", imageData, dx, dy, dirtyX, dirtyY, dirtyWidth, dirtyHeight)
}

// State

func (ctx *CanvasRenderingContext2D) Save() {
	ctx.Call("save")
}

func (ctx *CanvasRenderingContext2D) Restore() {
	ctx.Call("restore")
}

// TODO Hit regions:
// addHitRegion
// removeHitRegion
// clearHitRegions

type HTMLDListElement struct{ *BasicHTMLElement }

type HTMLDataElement struct {
	*BasicHTMLElement
}

func (e *HTMLDataElement) Value() string { return e.Get("value").String() }

func (e *HTMLDataElement) SetValue(v string) { e.Set("value", v) }

type HTMLDataListElement struct{ *BasicHTMLElement }

func (e *HTMLDataListElement) Options() []*HTMLOptionElement {
	return getOptions(e.Value, "options")
}

type HTMLDirectoryElement struct{ *BasicHTMLElement }
type HTMLDivElement struct{ *BasicHTMLElement }

type HTMLEmbedElement struct {
	*BasicHTMLElement
}

func (e *HTMLEmbedElement) Src() string  { return e.Get("src").String() }
func (e *HTMLEmbedElement) Type() string { return e.Get("type").String() }
func (e *HTMLEmbedElement) Width() int   { return e.Get("width").Int() }
func (e *HTMLEmbedElement) Height() int  { return e.Get("height").Int() }

func (e *HTMLEmbedElement) SetSrc(v string)  { e.Set("src", v) }
func (e *HTMLEmbedElement) SetType(v string) { e.Set("type", v) }
func (e *HTMLEmbedElement) SetWidth(v int)   { e.Set("width", v) }
func (e *HTMLEmbedElement) SetHeight(v int)  { e.Set("height", v) }

type HTMLFieldSetElement struct {
	*BasicHTMLElement
}

func (e *HTMLFieldSetElement) Disabled() bool            { return e.Get("disabled").Bool() }
func (e *HTMLFieldSetElement) Name() string              { return e.Get("name").String() }
func (e *HTMLFieldSetElement) Type() string              { return e.Get("type").String() }
func (e *HTMLFieldSetElement) ValidationMessage() string { return e.Get("validationMessage").String() }
func (e *HTMLFieldSetElement) WillValidate() bool        { return e.Get("willValidate").Bool() }

func (e *HTMLFieldSetElement) Elements() []HTMLElement {
	return nodeListToHTMLElements(e.Get("elements"))
}

func (e *HTMLFieldSetElement) Form() *HTMLFormElement {
	return getForm(e.Value)
}

func (e *HTMLFieldSetElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Get("validity")}
}

func (e *HTMLFieldSetElement) CheckValidity() bool {
	return e.Call("checkValidity").Bool()
}

func (e *HTMLFieldSetElement) SetCustomValidity(s string) {
	e.Call("setCustomValidity", s)
}

func (e *HTMLFieldSetElement) SetDisabled(v bool)            { e.Set("disabled", v) }
func (e *HTMLFieldSetElement) SetName(v string)              { e.Set("name", v) }
func (e *HTMLFieldSetElement) SetValidationMessage(v string) { e.Set("validationMessage", v) }

type HTMLFontElement struct{ *BasicHTMLElement }

type HTMLFormElement struct {
	*BasicHTMLElement
}

func (e *HTMLFormElement) AcceptCharset() string   { return e.Get("acceptCharset").String() }
func (e *HTMLFormElement) Action() string          { return e.Get("action").String() }
func (e *HTMLFormElement) Autocomplete() string    { return e.Get("autocomplete").String() }
func (e *HTMLFormElement) Encoding() string        { return e.Get("encoding").String() }
func (e *HTMLFormElement) Enctype() string         { return e.Get("enctype").String() }
func (e *HTMLFormElement) Length() int             { return e.Get("length").Int() }
func (e *HTMLFormElement) Method() string          { return e.Get("method").String() }
func (e *HTMLFormElement) Name() string            { return e.Get("name").String() }
func (e *HTMLFormElement) NoValidate() bool        { return e.Get("noValidate").Bool() }
func (e *HTMLFormElement) Target() string          { return e.Get("target").String() }
func (e *HTMLFormElement) Elements() []HTMLElement { return nodeListToHTMLElements(e.Get("elements")) }

func (e *HTMLFormElement) SetAcceptCharset(v string) { e.Set("acceptCharset", v) }
func (e *HTMLFormElement) SetAction(v string)        { e.Set("action", v) }
func (e *HTMLFormElement) SetAutocomplete(v string)  { e.Set("autocomplete", v) }
func (e *HTMLFormElement) SetEncoding(v string)      { e.Set("encoding", v) }
func (e *HTMLFormElement) SetEnctype(v string)       { e.Set("enctype", v) }
func (e *HTMLFormElement) SetMethod(v string)        { e.Set("method", v) }
func (e *HTMLFormElement) SetName(v string)          { e.Set("name", v) }
func (e *HTMLFormElement) SetNoValidate(v bool)      { e.Set("noValidate", v) }
func (e *HTMLFormElement) SetTarget(v string)        { e.Set("target", v) }

func (e *HTMLFormElement) CheckValidity() bool        { return e.Call("checkValidity").Bool() }
func (e *HTMLFormElement) Submit()                    { e.Call("submit") }
func (e *HTMLFormElement) Reset()                     { e.Call("reset") }
func (e *HTMLFormElement) Item(index int) HTMLElement { return wrapHTMLElement(e.Call("item", index)) }
func (e *HTMLFormElement) NamedItem(name string) HTMLElement {
	return wrapHTMLElement(e.Call("namedItem", name))
}

type HTMLFrameElement struct{ *BasicHTMLElement }
type HTMLFrameSetElement struct{ *BasicHTMLElement }
type HTMLHRElement struct{ *BasicHTMLElement }
type HTMLHeadElement struct{ *BasicHTMLElement }
type HTMLHeadingElement struct{ *BasicHTMLElement }
type HTMLHtmlElement struct{ *BasicHTMLElement }

type HTMLIFrameElement struct {
	*BasicHTMLElement
	// TODO sandbox attribute
}

func (e *HTMLIFrameElement) ContentDocument() Document { return wrapDocument(e.Get("contentDocument")) }
func (e *HTMLIFrameElement) ContentWindow() Window     { return &window{e.Get("contentWindow")} }
func (e *HTMLIFrameElement) Height() string            { return e.Get("height").String() }
func (e *HTMLIFrameElement) Name() string              { return e.Get("name").String() }
func (e *HTMLIFrameElement) Seamless() bool            { return e.Get("seamless").Bool() }
func (e *HTMLIFrameElement) Src() string               { return e.Get("src").String() }
func (e *HTMLIFrameElement) SrcDoc() string            { return e.Get("srcdoc").String() }
func (e *HTMLIFrameElement) Width() string             { return e.Get("width").String() }

func (e *HTMLIFrameElement) SetWidth(v string)  { e.Set("width", v) }
func (e *HTMLIFrameElement) SetHeight(v string) { e.Set("height", v) }
func (e *HTMLIFrameElement) SetName(v string)   { e.Set("name", v) }
func (e *HTMLIFrameElement) SetSrc(v string)    { e.Set("src", v) }
func (e *HTMLIFrameElement) SetSrcDoc(v string) { e.Set("srcdoc", v) }
func (e *HTMLIFrameElement) SetSeamless(v bool) { e.Set("seamless", v) }

type HTMLImageElement struct {
	*BasicHTMLElement
	// TODO constructor
}

func (e *HTMLImageElement) Complete() bool      { return e.Get("complete").Bool() }
func (e *HTMLImageElement) CrossOrigin() string { return e.Get("crossOrigin").String() }
func (e *HTMLImageElement) Height() int         { return e.Get("height").Int() }
func (e *HTMLImageElement) IsMap() bool         { return e.Get("isMap").Bool() }
func (e *HTMLImageElement) NaturalHeight() int  { return e.Get("naturalHeight").Int() }
func (e *HTMLImageElement) NaturalWidth() int   { return e.Get("naturalWidth").Int() }
func (e *HTMLImageElement) Src() string         { return e.Get("src").String() }
func (e *HTMLImageElement) UseMap() string      { return e.Get("useMap").String() }
func (e *HTMLImageElement) Width() int          { return e.Get("width").Int() }

func (e *HTMLImageElement) SetCrossOrigin(v string) { e.Set("crossOrigin", v) }
func (e *HTMLImageElement) SetHeight(v int)         { e.Set("height", v) }
func (e *HTMLImageElement) SetIsMap(v bool)         { e.Set("isMap", v) }
func (e *HTMLImageElement) SetSrc(v string)         { e.Set("src", v) }
func (e *HTMLImageElement) SetUseMap(v string)      { e.Set("useMap", v) }
func (e *HTMLImageElement) SetWidth(v int)          { e.Set("width", v) }

type HTMLInputElement struct {
	*BasicHTMLElement
}

func (e *HTMLInputElement) Accept() string             { return e.Get("accept").String() }
func (e *HTMLInputElement) Alt() string                { return e.Get("alt").String() }
func (e *HTMLInputElement) Autocomplete() string       { return e.Get("autocomplete").String() }
func (e *HTMLInputElement) Autofocus() bool            { return e.Get("autofocus").Bool() }
func (e *HTMLInputElement) Checked() bool              { return e.Get("checked").Bool() }
func (e *HTMLInputElement) DefaultChecked() bool       { return e.Get("defaultChecked").Bool() }
func (e *HTMLInputElement) DefaultValue() string       { return e.Get("defaultValue").String() }
func (e *HTMLInputElement) DirName() string            { return e.Get("dirName").String() }
func (e *HTMLInputElement) Disabled() bool             { return e.Get("disabled").Bool() }
func (e *HTMLInputElement) FormAction() string         { return e.Get("formAction").String() }
func (e *HTMLInputElement) FormEncType() string        { return e.Get("formEncType").String() }
func (e *HTMLInputElement) FormMethod() string         { return e.Get("formMethod").String() }
func (e *HTMLInputElement) FormNoValidate() bool       { return e.Get("formNoValidate").Bool() }
func (e *HTMLInputElement) FormTarget() string         { return e.Get("formTarget").String() }
func (e *HTMLInputElement) Height() string             { return e.Get("height").String() }
func (e *HTMLInputElement) Indeterminate() bool        { return e.Get("indeterminate").Bool() }
func (e *HTMLInputElement) Max() string                { return e.Get("max").String() }
func (e *HTMLInputElement) MaxLength() int             { return e.Get("maxLength").Int() }
func (e *HTMLInputElement) Min() string                { return e.Get("min").String() }
func (e *HTMLInputElement) Multiple() bool             { return e.Get("multiple").Bool() }
func (e *HTMLInputElement) Name() string               { return e.Get("name").String() }
func (e *HTMLInputElement) Pattern() string            { return e.Get("pattern").String() }
func (e *HTMLInputElement) Placeholder() string        { return e.Get("placeholder").String() }
func (e *HTMLInputElement) ReadOnly() bool             { return e.Get("readOnly").Bool() }
func (e *HTMLInputElement) Required() bool             { return e.Get("required").Bool() }
func (e *HTMLInputElement) SelectionDirection() string { return e.Get("selectionDirection").String() }
func (e *HTMLInputElement) SelectionEnd() int          { return e.Get("selectionEnd").Int() }
func (e *HTMLInputElement) SelectionStart() int        { return e.Get("selectionStart").Int() }
func (e *HTMLInputElement) Size() int                  { return e.Get("size").Int() }
func (e *HTMLInputElement) Src() string                { return e.Get("src").String() }
func (e *HTMLInputElement) Step() string               { return e.Get("step").String() }
func (e *HTMLInputElement) TabIndex() int              { return e.Get("tabIndex").Int() }
func (e *HTMLInputElement) Type() string               { return e.Get("type").String() }
func (e *HTMLInputElement) ValidationMessage() string  { return e.Get("validationMessage").String() }
func (e *HTMLInputElement) Value() string              { return e.Get("value").String() }
func (e *HTMLInputElement) ValueAsDate() time.Time {
	return time.Unix(0, int64(e.Get("valueAsDate").Call("getTime").Int())*1000000)
}
func (e *HTMLInputElement) ValueAsNumber() float64 { return e.Get("valueAsNumber").Float() }
func (e *HTMLInputElement) Width() string          { return e.Get("width").String() }
func (e *HTMLInputElement) WillValidate() bool     { return e.Get("willValidate").Bool() }

func (e *HTMLInputElement) SetAccept(v string)             { e.Set("accept", v) }
func (e *HTMLInputElement) SetAlt(v string)                { e.Set("alt", v) }
func (e *HTMLInputElement) SetAutocomplete(v string)       { e.Set("autocomplete", v) }
func (e *HTMLInputElement) SetAutofocus(v bool)            { e.Set("autofocus", v) }
func (e *HTMLInputElement) SetChecked(v bool)              { e.Set("checked", v) }
func (e *HTMLInputElement) SetDefaultChecked(v bool)       { e.Set("defaultChecked", v) }
func (e *HTMLInputElement) SetDefaultValue(v string)       { e.Set("defaultValue", v) }
func (e *HTMLInputElement) SetDirName(v string)            { e.Set("dirName", v) }
func (e *HTMLInputElement) SetDisabled(v bool)             { e.Set("disabled", v) }
func (e *HTMLInputElement) SetFormAction(v string)         { e.Set("formAction", v) }
func (e *HTMLInputElement) SetFormEncType(v string)        { e.Set("formEncType", v) }
func (e *HTMLInputElement) SetFormMethod(v string)         { e.Set("formMethod", v) }
func (e *HTMLInputElement) SetFormNoValidate(v bool)       { e.Set("formNoValidate", v) }
func (e *HTMLInputElement) SetFormTarget(v string)         { e.Set("formTarget", v) }
func (e *HTMLInputElement) SetHeight(v string)             { e.Set("height", v) }
func (e *HTMLInputElement) SetMax(v string)                { e.Set("max", v) }
func (e *HTMLInputElement) SetMaxLength(v int)             { e.Set("maxLength", v) }
func (e *HTMLInputElement) SetMin(v string)                { e.Set("min", v) }
func (e *HTMLInputElement) SetMultiple(v bool)             { e.Set("multiple", v) }
func (e *HTMLInputElement) SetName(v string)               { e.Set("name", v) }
func (e *HTMLInputElement) SetPattern(v string)            { e.Set("pattern", v) }
func (e *HTMLInputElement) SetPlaceholder(v string)        { e.Set("placeholder", v) }
func (e *HTMLInputElement) SetReadOnly(v bool)             { e.Set("readOnly", v) }
func (e *HTMLInputElement) SetRequired(v bool)             { e.Set("required", v) }
func (e *HTMLInputElement) SetSelectionDirection(v string) { e.Set("selectionDirection", v) }
func (e *HTMLInputElement) SetSelectionEnd(v int)          { e.Set("selectionEnd", v) }
func (e *HTMLInputElement) SetSelectionStart(v int)        { e.Set("selectionStart", v) }
func (e *HTMLInputElement) SetSize(v int)                  { e.Set("size", v) }
func (e *HTMLInputElement) SetSrc(v string)                { e.Set("src", v) }
func (e *HTMLInputElement) SetStep(v string)               { e.Set("step", v) }
func (e *HTMLInputElement) SetType(v string)               { e.Set("type", v) }
func (e *HTMLInputElement) SetValue(v string)              { e.Set("value", v) }
func (e *HTMLInputElement) SetWidth(v string)              { e.Set("width", v) }
func (e *HTMLInputElement) SetWillValidate(v bool)         { e.Set("willValidate", v) }

// File represents files as can be obtained from file choosers or drag
// and drop. The dom package does not define any methods on File nor
// does it provide access to the blob or a way to read it.
type File struct {
	js.Value
}

func (e *HTMLInputElement) Files() []*File {
	files := e.Get("files")
	out := make([]*File, files.Get("length").Int())
	for i := range out {
		out[i] = &File{files.Call("item", i)}
	}
	return out
}

func (e *HTMLInputElement) List() *HTMLDataListElement {
	list := wrapHTMLElement(e.Get("list"))
	if list == nil {
		return nil
	}
	return list.(*HTMLDataListElement)
}

func (e *HTMLInputElement) Labels() []*HTMLLabelElement { return getLabels(e.Underlying()) }
func (e *HTMLInputElement) Form() *HTMLFormElement      { return getForm(e.Underlying()) }
func (e *HTMLInputElement) Validity() *ValidityState    { return &ValidityState{Value: e.Get("validity")} }
func (e *HTMLInputElement) CheckValidity() bool         { return e.Call("checkValidity").Bool() }
func (e *HTMLInputElement) SetCustomValidity(s string)  { e.Call("setCustomValidity", s) }
func (e *HTMLInputElement) Select()                     { e.Call("select") }
func (e *HTMLInputElement) SetSelectionRange(start, end int, direction string) {
	e.Call("setSelectionRange", start, end, direction)
}
func (e *HTMLInputElement) StepDown(n int) error { return callRecover(e.Underlying(), "stepDown", n) }
func (e *HTMLInputElement) StepUp(n int) error   { return callRecover(e.Underlying(), "stepUp", n) }

type HTMLKeygenElement struct {
	*BasicHTMLElement
}

func (e *HTMLKeygenElement) Autofocus() bool             { return e.Get("autofocus").Bool() }
func (e *HTMLKeygenElement) Challenge() string           { return e.Get("challenge").String() }
func (e *HTMLKeygenElement) Disabled() bool              { return e.Get("disabled").Bool() }
func (e *HTMLKeygenElement) Form() *HTMLFormElement      { return getForm(e.Value) }
func (e *HTMLKeygenElement) Keytype() string             { return e.Get("keytype").String() }
func (e *HTMLKeygenElement) Labels() []*HTMLLabelElement { return getLabels(e.Value) }
func (e *HTMLKeygenElement) Name() string                { return e.Get("name").String() }
func (e *HTMLKeygenElement) Type() string                { return e.Get("type").String() }
func (e *HTMLKeygenElement) ValidationMessage() string   { return e.Get("validationMessage").String() }
func (e *HTMLKeygenElement) Validity() *ValidityState    { return &ValidityState{Value: e.Get("validity")} }
func (e *HTMLKeygenElement) WillValidate() bool          { return e.Get("willValidate").Bool() }

func (e *HTMLKeygenElement) SetAutofocus(v bool)        { e.Set("autofocus", v) }
func (e *HTMLKeygenElement) SetChallenge(v string)      { e.Set("challenge", v) }
func (e *HTMLKeygenElement) SetDisabled(v bool)         { e.Set("disabled", v) }
func (e *HTMLKeygenElement) SetKeytype(v string)        { e.Set("keytype", v) }
func (e *HTMLKeygenElement) SetName(v string)           { e.Set("name", v) }
func (e *HTMLKeygenElement) SetCustomValidity(s string) { e.Call("setCustomValidity", s) }

func (e *HTMLKeygenElement) CheckValidity() bool { return e.Call("checkValidity").Bool() }

type HTMLLIElement struct {
	*BasicHTMLElement
}

func (e *HTMLLIElement) Value() int { return e.Get("value").Int() }

type HTMLLabelElement struct {
	*BasicHTMLElement
}

func (e *HTMLLabelElement) For() string            { return e.Get("htmlFor").String() }
func (e *HTMLLabelElement) Control() HTMLElement   { return wrapHTMLElement(e.Get("control")) }
func (e *HTMLLabelElement) Form() *HTMLFormElement { return getForm(e.Value) }

type HTMLLegendElement struct{ *BasicHTMLElement }

func (e *HTMLLegendElement) Form() *HTMLFormElement { return getForm(e.Value) }

type HTMLLinkElement struct {
	*BasicHTMLElement
}

func (e *HTMLLinkElement) Disabled() bool   { return e.Get("disabled").Bool() }
func (e *HTMLLinkElement) Href() string     { return e.Get("href").String() }
func (e *HTMLLinkElement) HrefLang() string { return e.Get("hrefLang").String() }
func (e *HTMLLinkElement) Media() string    { return e.Get("media").String() }
func (e *HTMLLinkElement) Type() string     { return e.Get("type").String() }

func (e *HTMLLinkElement) SetDisabled(v bool)   { e.Set("disabled", v) }
func (e *HTMLLinkElement) SetHref(v string)     { e.Set("href", v) }
func (e *HTMLLinkElement) SetHrefLang(v string) { e.Set("hrefLang", v) }
func (e *HTMLLinkElement) SetMedia(v string)    { e.Set("media", v) }
func (e *HTMLLinkElement) SetType(v string)     { e.Set("type", v) }

func (e *HTMLLinkElement) Rel() *TokenList {
	return &TokenList{dtl: e.Get("relList"), o: e.Value, sa: "rel"}
}

func (e *HTMLLinkElement) Sizes() *TokenList {
	return &TokenList{dtl: e.Get("sizes"), o: e.Value}
}

func (e *HTMLLinkElement) Sheet() StyleSheet {
	// FIXME implement
	panic("not implemented")
}

type HTMLMapElement struct {
	*BasicHTMLElement
}

func (e *HTMLMapElement) Name() string { return e.Get("name").String() }

func (e *HTMLMapElement) SetName(v string) { e.Set("name", v) }

func (e *HTMLMapElement) Areas() []*HTMLAreaElement {
	areas := nodeListToElements(e.Get("areas"))
	out := make([]*HTMLAreaElement, len(areas))
	for i, area := range areas {
		out[i] = area.(*HTMLAreaElement)
	}
	return out
}

func (e *HTMLMapElement) Images() []HTMLElement {
	return nodeListToHTMLElements(e.Get("areas"))
}

type HTMLMediaElement struct {
	*BasicHTMLElement
}

func (e *HTMLMediaElement) Paused() bool { return e.Get("paused").Bool() }

func (e *HTMLMediaElement) SetPaused(v bool) { e.Set("paused", v) }

func (e *HTMLMediaElement) Play()  { e.Call("play") }
func (e *HTMLMediaElement) Pause() { e.Call("pause") }

type HTMLMenuElement struct{ *BasicHTMLElement }

type HTMLMetaElement struct {
	*BasicHTMLElement
}

func (e *HTMLMetaElement) Content() string   { return e.Get("content").String() }
func (e *HTMLMetaElement) HTTPEquiv() string { return e.Get("httpEquiv").String() }
func (e *HTMLMetaElement) Name() string      { return e.Get("name").String() }

func (e *HTMLMetaElement) SetContent(v string)   { e.Set("content", v) }
func (e *HTMLMetaElement) SetHTTPEquiv(v string) { e.Set("httpEquiv", v) }
func (e *HTMLMetaElement) SetName(v string)      { e.Set("name", v) }

type HTMLMeterElement struct {
	*BasicHTMLElement
}

func (e *HTMLMeterElement) High() float64              { return e.Get("high").Float() }
func (e *HTMLMeterElement) Low() float64               { return e.Get("low").Float() }
func (e *HTMLMeterElement) Max() float64               { return e.Get("max").Float() }
func (e *HTMLMeterElement) Min() float64               { return e.Get("min").Float() }
func (e *HTMLMeterElement) Optimum() float64           { return e.Get("optimum").Float() }
func (e HTMLMeterElement) Labels() []*HTMLLabelElement { return getLabels(e.Value) }

func (e *HTMLMeterElement) SetHigh(v float64)    { e.Set("high", v) }
func (e *HTMLMeterElement) SetLow(v float64)     { e.Set("low", v) }
func (e *HTMLMeterElement) SetMax(v float64)     { e.Set("max", v) }
func (e *HTMLMeterElement) SetMin(v float64)     { e.Set("min", v) }
func (e *HTMLMeterElement) SetOptimum(v float64) { e.Set("optimum", v) }

type HTMLModElement struct {
	*BasicHTMLElement
}

func (e *HTMLModElement) Cite() string     { return e.Get("cite").String() }
func (e *HTMLModElement) DateTime() string { return e.Get("dateTime").String() }

func (e *HTMLModElement) SetCite(v string)     { e.Set("cite", v) }
func (e *HTMLModElement) SetDateTime(v string) { e.Set("dateTime", v) }

type HTMLOListElement struct {
	*BasicHTMLElement
}

func (e *HTMLOListElement) Reversed() bool { return e.Get("reversed").Bool() }
func (e *HTMLOListElement) Start() int     { return e.Get("start").Int() }
func (e *HTMLOListElement) Type() string   { return e.Get("type").String() }

func (e *HTMLOListElement) SetReversed(v bool) { e.Set("reversed", v) }
func (e *HTMLOListElement) SetStart(v int)     { e.Set("start", v) }
func (e *HTMLOListElement) SetType(v string)   { e.Set("type", v) }

type HTMLObjectElement struct {
	*BasicHTMLElement
}

func (e *HTMLObjectElement) CheckValidity() bool       { return e.Call("checkValidity").Bool() }
func (e *HTMLObjectElement) ContentDocument() Document { return wrapDocument(e.Get("contentDocument")) }
func (e *HTMLObjectElement) ContentWindow() Window     { return &window{e.Get("contentWindow")} }
func (e *HTMLObjectElement) Data() string              { return e.Get("data").String() }
func (e *HTMLObjectElement) Form() *HTMLFormElement    { return getForm(e.Value) }
func (e *HTMLObjectElement) Height() string            { return e.Get("height").String() }
func (e *HTMLObjectElement) Name() string              { return e.Get("name").String() }
func (e *HTMLObjectElement) TabIndex() int             { return e.Get("tabIndex").Int() }
func (e *HTMLObjectElement) Type() string              { return e.Get("type").String() }
func (e *HTMLObjectElement) TypeMustMatch() bool       { return e.Get("typeMustMatch").Bool() }
func (e *HTMLObjectElement) UseMap() string            { return e.Get("useMap").String() }
func (e *HTMLObjectElement) ValidationMessage() string { return e.Get("validationMessage").String() }
func (e *HTMLObjectElement) Validity() *ValidityState  { return &ValidityState{Value: e.Get("validity")} }
func (e *HTMLObjectElement) WillValidate() bool        { return e.Get("willValidate").Bool() }
func (e *HTMLObjectElement) With() string              { return e.Get("with").String() }

func (e *HTMLObjectElement) SetCustomValidity(s string) { e.Call("setCustomValidity", s) }
func (e *HTMLObjectElement) SetData(v string)           { e.Set("data", v) }
func (e *HTMLObjectElement) SetHeight(v string)         { e.Set("height", v) }
func (e *HTMLObjectElement) SetName(v string)           { e.Set("name", v) }
func (e *HTMLObjectElement) SetTabIndex(v int)          { e.Set("tabIndex", v) }
func (e *HTMLObjectElement) SetType(v string)           { e.Set("type", v) }
func (e *HTMLObjectElement) SetTypeMustMatch(v bool)    { e.Set("typeMustMatch", v) }
func (e *HTMLObjectElement) SetUseMap(v string)         { e.Set("useMap", v) }
func (e *HTMLObjectElement) SetWith(v string)           { e.Set("with", v) }

type HTMLOptGroupElement struct {
	*BasicHTMLElement
}

func (e *HTMLOptGroupElement) Disabled() bool { return e.Get("disabled").Bool() }
func (e *HTMLOptGroupElement) Label() string  { return e.Get("label").String() }

func (e *HTMLOptGroupElement) SetDisabled(v bool) { e.Set("disabled", v) }
func (e *HTMLOptGroupElement) SetLabel(v string)  { e.Set("label", v) }

type HTMLOptionElement struct {
	*BasicHTMLElement
}

func (e *HTMLOptionElement) DefaultSelected() bool  { return e.Get("defaultSelected").Bool() }
func (e *HTMLOptionElement) Disabled() bool         { return e.Get("disabled").Bool() }
func (e *HTMLOptionElement) Form() *HTMLFormElement { return getForm(e.Underlying()) }
func (e *HTMLOptionElement) Index() int             { return e.Get("index").Int() }
func (e *HTMLOptionElement) Label() string          { return e.Get("label").String() }
func (e *HTMLOptionElement) Selected() bool         { return e.Get("selected").Bool() }
func (e *HTMLOptionElement) Text() string           { return e.Get("text").String() }
func (e *HTMLOptionElement) Value() string          { return e.Get("value").String() }

func (e *HTMLOptionElement) SetDefaultSelected(v bool) { e.Set("defaultSelected", v) }
func (e *HTMLOptionElement) SetDisabled(v bool)        { e.Set("disabled", v) }
func (e *HTMLOptionElement) SetLabel(v string)         { e.Set("label", v) }
func (e *HTMLOptionElement) SetSelected(v bool)        { e.Set("selected", v) }
func (e *HTMLOptionElement) SetText(v string)          { e.Set("text", v) }
func (e *HTMLOptionElement) SetValue(v string)         { e.Set("value", v) }

type HTMLOutputElement struct {
	*BasicHTMLElement
}

func (e *HTMLOutputElement) DefaultValue() string        { return e.Get("defaultValue").String() }
func (e *HTMLOutputElement) Name() string                { return e.Get("name").String() }
func (e *HTMLOutputElement) Type() string                { return e.Get("type").String() }
func (e *HTMLOutputElement) ValidationMessage() string   { return e.Get("validationMessage").String() }
func (e *HTMLOutputElement) Value() string               { return e.Get("value").String() }
func (e *HTMLOutputElement) WillValidate() bool          { return e.Get("willValidate").Bool() }
func (e *HTMLOutputElement) Form() *HTMLFormElement      { return getForm(e.Underlying()) }
func (e *HTMLOutputElement) Labels() []*HTMLLabelElement { return getLabels(e.Underlying()) }
func (e *HTMLOutputElement) Validity() *ValidityState    { return &ValidityState{Value: e.Get("validity")} }
func (e *HTMLOutputElement) For() *TokenList {
	return &TokenList{dtl: e.Get("htmlFor"), o: e.Underlying()}
}
func (e *HTMLOutputElement) CheckValidity() bool        { return e.Call("checkValidity").Bool() }
func (e *HTMLOutputElement) SetCustomValidity(s string) { e.Call("setCustomValidity", s) }

func (e *HTMLOutputElement) SetDefaultValue(v string) { e.Set("defaultValue", v) }
func (e *HTMLOutputElement) SetName(v string)         { e.Set("name", v) }
func (e *HTMLOutputElement) SetValue(v string)        { e.Set("value", v) }

type HTMLParagraphElement struct{ *BasicHTMLElement }

type HTMLParamElement struct {
	*BasicHTMLElement
}

func (e *HTMLParamElement) Name() string  { return e.Get("name").String() }
func (e *HTMLParamElement) Value() string { return e.Get("value").String() }

func (e *HTMLParamElement) SetName(v string)  { e.Set("name", v) }
func (e *HTMLParamElement) SetValue(v string) { e.Set("value", v) }

type HTMLPreElement struct{ *BasicHTMLElement }

type HTMLProgressElement struct {
	*BasicHTMLElement
}

func (e *HTMLProgressElement) Max() float64                { return e.Get("max").Float() }
func (e *HTMLProgressElement) Position() float64           { return e.Get("position").Float() }
func (e *HTMLProgressElement) Value() float64              { return e.Get("value").Float() }
func (e *HTMLProgressElement) Labels() []*HTMLLabelElement { return getLabels(e.Underlying()) }

func (e *HTMLProgressElement) SetMax(v float64)   { e.Set("max", v) }
func (e *HTMLProgressElement) SetValue(v float64) { e.Set("value", v) }

type HTMLQuoteElement struct {
	*BasicHTMLElement
}

func (e *HTMLQuoteElement) Cite() string { return e.Get("cite").String() }

type HTMLScriptElement struct {
	*BasicHTMLElement
}

func (e *HTMLScriptElement) Type() string    { return e.Get("type").String() }
func (e *HTMLScriptElement) Src() string     { return e.Get("src").String() }
func (e *HTMLScriptElement) Charset() string { return e.Get("charset").String() }
func (e *HTMLScriptElement) Async() bool     { return e.Get("async").Bool() }
func (e *HTMLScriptElement) Defer() bool     { return e.Get("defer").Bool() }
func (e *HTMLScriptElement) Text() string    { return e.Get("text").String() }

func (e *HTMLScriptElement) SetType(v string)    { e.Set("type", v) }
func (e *HTMLScriptElement) SetSrc(v string)     { e.Set("src", v) }
func (e *HTMLScriptElement) SetCharset(v string) { e.Set("charset", v) }
func (e *HTMLScriptElement) SetAsync(v bool)     { e.Set("async", v) }
func (e *HTMLScriptElement) SetDefer(v bool)     { e.Set("defer", v) }
func (e *HTMLScriptElement) SetText(v string)    { e.Set("text", v) }

type HTMLSelectElement struct {
	*BasicHTMLElement
}

func (e *HTMLSelectElement) Validity() *ValidityState    { return &ValidityState{Value: e.Get("validity")} }
func (e *HTMLSelectElement) Autofocus() bool             { return e.Get("autofocus").Bool() }
func (e *HTMLSelectElement) Disabled() bool              { return e.Get("disabled").Bool() }
func (e *HTMLSelectElement) Length() int                 { return e.Get("length").Int() }
func (e *HTMLSelectElement) Multiple() bool              { return e.Get("multiple").Bool() }
func (e *HTMLSelectElement) Name() string                { return e.Get("name").String() }
func (e *HTMLSelectElement) Required() bool              { return e.Get("required").Bool() }
func (e *HTMLSelectElement) SelectedIndex() int          { return e.Get("selectedIndex").Int() }
func (e *HTMLSelectElement) Size() int                   { return e.Get("size").Int() }
func (e *HTMLSelectElement) Type() string                { return e.Get("type").String() }
func (e *HTMLSelectElement) ValidationMessage() string   { return e.Get("validationMessage").String() }
func (e *HTMLSelectElement) Value() string               { return e.Get("value").String() }
func (e *HTMLSelectElement) WillValidate() bool          { return e.Get("willValidate").Bool() }
func (e *HTMLSelectElement) Labels() []*HTMLLabelElement { return getLabels(e.Underlying()) }
func (e *HTMLSelectElement) Form() *HTMLFormElement      { return getForm(e.Underlying()) }
func (e *HTMLSelectElement) Options() []*HTMLOptionElement {
	return getOptions(e.Underlying(), "options")
}
func (e *HTMLSelectElement) SelectedOptions() []*HTMLOptionElement {
	return getOptions(e.Underlying(), "selectedOptions")
}
func (e *HTMLSelectElement) Item(index int) *HTMLOptionElement {
	el := wrapHTMLElement(e.Call("item", index))
	if el == nil {
		return nil
	}
	return el.(*HTMLOptionElement)
}
func (e *HTMLSelectElement) NamedItem(name string) *HTMLOptionElement {
	el := wrapHTMLElement(e.Call("namedItem", name))
	if el == nil {
		return nil
	}
	return el.(*HTMLOptionElement)
}

func (e *HTMLSelectElement) SetAutofocus(v bool)        { e.Set("autofocus", v) }
func (e *HTMLSelectElement) SetDisabled(v bool)         { e.Set("disabled", v) }
func (e *HTMLSelectElement) SetLength(v int)            { e.Set("length", v) }
func (e *HTMLSelectElement) SetMultiple(v bool)         { e.Set("multiple", v) }
func (e *HTMLSelectElement) SetName(v string)           { e.Set("name", v) }
func (e *HTMLSelectElement) SetRequired(v bool)         { e.Set("required", v) }
func (e *HTMLSelectElement) SetSelectedIndex(v int)     { e.Set("selectedIndex", v) }
func (e *HTMLSelectElement) SetSize(v int)              { e.Set("size", v) }
func (e *HTMLSelectElement) SetValue(v string)          { e.Set("value", v) }
func (e *HTMLSelectElement) SetCustomValidity(s string) { e.Call("setCustomValidity", s) }

func (e *HTMLSelectElement) CheckValidity() bool { return e.Call("checkValidity").Bool() }

// TODO(dominikh): Not implementing Add or Remove for now. For one,
// Add with "before" behaves weird when dealing with optgroups. Also,
// there's already InsertBefore and RemoveChild which can be used
// instead.

type HTMLSourceElement struct {
	*BasicHTMLElement
}

func (e *HTMLSourceElement) Media() string { return e.Get("media").String() }
func (e *HTMLSourceElement) Src() string   { return e.Get("src").String() }
func (e *HTMLSourceElement) Type() string  { return e.Get("type").String() }

func (e *HTMLSourceElement) SetMedia(v string) { e.Set("media", v) }
func (e *HTMLSourceElement) SetSrc(v string)   { e.Set("src", v) }
func (e *HTMLSourceElement) SetType(v string)  { e.Set("type", v) }

type HTMLSpanElement struct{ *BasicHTMLElement }
type HTMLStyleElement struct{ *BasicHTMLElement }
type HTMLTableCaptionElement struct{ *BasicHTMLElement }

type HTMLTableCellElement struct {
	*BasicHTMLElement
	// TODO headers
}

func (e *HTMLTableCellElement) ColSpan() int   { return e.Get("colSpan").Int() }
func (e *HTMLTableCellElement) RowSpan() int   { return e.Get("rowSpan").Int() }
func (e *HTMLTableCellElement) CellIndex() int { return e.Get("cellIndex").Int() }

func (e *HTMLTableCellElement) SetColSpan(v int) { e.Set("colSpan", v) }
func (e *HTMLTableCellElement) SetRowSpan(v int) { e.Set("rowSpan", v) }

type HTMLTableColElement struct {
	*BasicHTMLElement
}

func (e *HTMLTableColElement) Span() int { return e.Get("span").Int() }

func (e *HTMLTableColElement) SetSpan(v int) { e.Set("span", v) }

type HTMLTableDataCellElement struct{ *BasicHTMLElement }
type HTMLTableElement struct{ *BasicHTMLElement }

type HTMLTableHeaderCellElement struct {
	*BasicHTMLElement
}

func (e *HTMLTableHeaderCellElement) Abbr() string  { return e.Get("abbr").String() }
func (e *HTMLTableHeaderCellElement) Scope() string { return e.Get("scope").String() }

func (e *HTMLTableHeaderCellElement) SetAbbr(v string)  { e.Set("abbr", v) }
func (e *HTMLTableHeaderCellElement) SetScope(v string) { e.Set("scope", v) }

type HTMLTableRowElement struct {
	*BasicHTMLElement
}

func (e *HTMLTableRowElement) RowIndex() int        { return e.Get("rowIndex").Int() }
func (e *HTMLTableRowElement) SectionRowIndex() int { return e.Get("sectionRowIndex").Int() }
func (e *HTMLTableRowElement) Cells() []*HTMLTableCellElement {
	cells := nodeListToElements(e.Get("cells"))
	out := make([]*HTMLTableCellElement, len(cells))
	for i, cell := range cells {
		out[i] = cell.(*HTMLTableCellElement)
	}
	return out
}

func (e *HTMLTableRowElement) InsertCell(index int) *HTMLTableCellElement {
	return wrapHTMLElement(e.Call("insertCell", index)).(*HTMLTableCellElement)
}

func (e *HTMLTableRowElement) DeleteCell(index int) {
	// FIXME exception handling/check that index is in bounds
	e.Call("deleteCell", index)
}

type HTMLTableSectionElement struct{ *BasicHTMLElement }

func (e *HTMLTableSectionElement) Rows() []*HTMLTableRowElement {
	rows := nodeListToElements(e.Get("rows"))
	out := make([]*HTMLTableRowElement, len(rows))
	for i, row := range rows {
		out[i] = row.(*HTMLTableRowElement)
	}
	return out
}

func (e *HTMLTableSectionElement) DeleteRow(index int) {
	// FIXME exception handling/check that index is in bounds
	e.Call("deleteRow", index)
}

func (e *HTMLTableSectionElement) InsertRow(index int) *HTMLTableRowElement {
	return wrapHTMLElement(e.Call("insertRow", index)).(*HTMLTableRowElement)
}

type HTMLTemplateElement struct{ *BasicHTMLElement }

func (e *HTMLTemplateElement) Content() DocumentFragment {
	return wrapDocumentFragment(e.Get("content"))
}

type HTMLTextAreaElement struct {
	*BasicHTMLElement
}

func (e *HTMLTextAreaElement) Autocomplete() string        { return e.Get("autocomplete").String() }
func (e *HTMLTextAreaElement) Autofocus() bool             { return e.Get("autofocus").Bool() }
func (e *HTMLTextAreaElement) Cols() int                   { return e.Get("cols").Int() }
func (e *HTMLTextAreaElement) DefaultValue() string        { return e.Get("defaultValue").String() }
func (e *HTMLTextAreaElement) DirName() string             { return e.Get("dirName").String() }
func (e *HTMLTextAreaElement) Disabled() bool              { return e.Get("disabled").Bool() }
func (e *HTMLTextAreaElement) MaxLength() int              { return e.Get("maxLength").Int() }
func (e *HTMLTextAreaElement) Name() string                { return e.Get("name").String() }
func (e *HTMLTextAreaElement) Placeholder() string         { return e.Get("placeholder").String() }
func (e *HTMLTextAreaElement) ReadOnly() bool              { return e.Get("readOnly").Bool() }
func (e *HTMLTextAreaElement) Required() bool              { return e.Get("required").Bool() }
func (e *HTMLTextAreaElement) Rows() int                   { return e.Get("rows").Int() }
func (e *HTMLTextAreaElement) SelectionDirection() string  { return e.Get("selectionDirection").String() }
func (e *HTMLTextAreaElement) SelectionStart() int         { return e.Get("selectionStart").Int() }
func (e *HTMLTextAreaElement) SelectionEnd() int           { return e.Get("selectionEnd").Int() }
func (e *HTMLTextAreaElement) TabIndex() int               { return e.Get("tabIndex").Int() }
func (e *HTMLTextAreaElement) TextLength() int             { return e.Get("textLength").Int() }
func (e *HTMLTextAreaElement) Type() string                { return e.Get("type").String() }
func (e *HTMLTextAreaElement) ValidationMessage() string   { return e.Get("validationMessage").String() }
func (e *HTMLTextAreaElement) Value() string               { return e.Get("value").String() }
func (e *HTMLTextAreaElement) WillValidate() bool          { return e.Get("willValidate").Bool() }
func (e *HTMLTextAreaElement) Wrap() string                { return e.Get("wrap").String() }
func (e *HTMLTextAreaElement) Form() *HTMLFormElement      { return getForm(e.Underlying()) }
func (e *HTMLTextAreaElement) Labels() []*HTMLLabelElement { return getLabels(e.Underlying()) }
func (e *HTMLTextAreaElement) Validity() *ValidityState {
	return &ValidityState{Value: e.Get("validity")}
}

func (e *HTMLTextAreaElement) SetAutocomplete(v string)       { e.Set("autocomplete", v) }
func (e *HTMLTextAreaElement) SetAutofocus(v bool)            { e.Set("autofocus", v) }
func (e *HTMLTextAreaElement) SetCols(v int)                  { e.Set("cols", v) }
func (e *HTMLTextAreaElement) SetDefaultValue(v string)       { e.Set("defaultValue", v) }
func (e *HTMLTextAreaElement) SetDirName(v string)            { e.Set("dirName", v) }
func (e *HTMLTextAreaElement) SetDisabled(v bool)             { e.Set("disabled", v) }
func (e *HTMLTextAreaElement) SetMaxLength(v int)             { e.Set("maxLength", v) }
func (e *HTMLTextAreaElement) SetName(v string)               { e.Set("name", v) }
func (e *HTMLTextAreaElement) SetPlaceholder(v string)        { e.Set("placeholder", v) }
func (e *HTMLTextAreaElement) SetReadOnly(v bool)             { e.Set("readOnly", v) }
func (e *HTMLTextAreaElement) SetRequired(v bool)             { e.Set("required", v) }
func (e *HTMLTextAreaElement) SetRows(v int)                  { e.Set("rows", v) }
func (e *HTMLTextAreaElement) SetSelectionDirection(v string) { e.Set("selectionDirection", v) }
func (e *HTMLTextAreaElement) SetSelectionStart(v int)        { e.Set("selectionStart", v) }
func (e *HTMLTextAreaElement) SetSelectionEnd(v int)          { e.Set("selectionEnd", v) }
func (e *HTMLTextAreaElement) SetTabIndex(v int)              { e.Set("tabIndex", v) }
func (e *HTMLTextAreaElement) SetValue(v string)              { e.Set("value", v) }
func (e *HTMLTextAreaElement) SetWrap(v string)               { e.Set("wrap", v) }
func (e *HTMLTextAreaElement) SetCustomValidity(s string)     { e.Call("setCustomValidity", s) }
func (e *HTMLTextAreaElement) SetSelectionRange(start, end int, direction string) {
	e.Call("setSelectionRange", start, end, direction)
}

func (e *HTMLTextAreaElement) CheckValidity() bool { return e.Call("checkValidity").Bool() }
func (e *HTMLTextAreaElement) Select()             { e.Call("select") }

type HTMLTimeElement struct {
	*BasicHTMLElement
}

func (e *HTMLTimeElement) DateTime() string { return e.Get("dateTime").String() }

type HTMLTitleElement struct {
	*BasicHTMLElement
}

func (e *HTMLTitleElement) Text() string { return e.Get("text").String() }

// TextTrack represents text track data for <track> elements. It does
// not currently provide any methods or attributes and it hasn't been
// decided yet whether they will be added to this package or a
// separate package.
type TextTrack struct{ js.Value }

type HTMLTrackElement struct {
	*BasicHTMLElement
}

func (e *HTMLTrackElement) Kind() string      { return e.Get("kind").String() }
func (e *HTMLTrackElement) Src() string       { return e.Get("src").String() }
func (e *HTMLTrackElement) Srclang() string   { return e.Get("srclang").String() }
func (e *HTMLTrackElement) Label() string     { return e.Get("label").String() }
func (e *HTMLTrackElement) Default() bool     { return e.Get("default").Bool() }
func (e *HTMLTrackElement) ReadyState() int   { return e.Get("readyState").Int() }
func (e *HTMLTrackElement) Track() *TextTrack { return &TextTrack{e.Get("track")} }

func (e *HTMLTrackElement) SetKind(v string)    { e.Set("kind", v) }
func (e *HTMLTrackElement) SetSrc(v string)     { e.Set("src", v) }
func (e *HTMLTrackElement) SetSrclang(v string) { e.Set("srclang", v) }
func (e *HTMLTrackElement) SetLabel(v string)   { e.Set("label", v) }
func (e *HTMLTrackElement) SetDefault(v bool)   { e.Set("default", v) }

type HTMLUListElement struct{ *BasicHTMLElement }
type HTMLUnknownElement struct{ *BasicHTMLElement }

type HTMLVideoElement struct{ *HTMLMediaElement }

type ValidityState struct {
	js.Value
}

func (s *ValidityState) CustomError() bool     { return s.Get("customError").Bool() }
func (s *ValidityState) PatternMismatch() bool { return s.Get("patternMismatch").Bool() }
func (s *ValidityState) RangeOverflow() bool   { return s.Get("rangeOverflow").Bool() }
func (s *ValidityState) RangeUnderflow() bool  { return s.Get("rangeUnderflow").Bool() }
func (s *ValidityState) StepMismatch() bool    { return s.Get("stepMismatch").Bool() }
func (s *ValidityState) TooLong() bool         { return s.Get("tooLong").Bool() }
func (s *ValidityState) TypeMismatch() bool    { return s.Get("typeMismatch").Bool() }
func (s *ValidityState) Valid() bool           { return s.Get("valid").Bool() }
func (s *ValidityState) ValueMissing() bool    { return s.Get("valueMissing").Bool() }

type CSSStyleDeclaration struct{ js.Value }

func (css *CSSStyleDeclaration) ToMap() map[string]string {
	m := make(map[string]string)
	N := css.Get("length").Int()
	for i := 0; i < N; i++ {
		name := css.Call("item", i).String()
		value := css.Call("getPropertyValue", name).String()
		m[name] = value
	}

	return m
}

func (css *CSSStyleDeclaration) RemoveProperty(name string) { css.Call("removeProperty", name) }

func (css *CSSStyleDeclaration) GetPropertyValue(name string) string {
	return toString(css.Call("getPropertyValue", name))
}

func (css *CSSStyleDeclaration) GetPropertyPriority(name string) string {
	return toString(css.Call("getPropertyPriority", name))
}

func (css *CSSStyleDeclaration) SetProperty(name, value, priority string) {
	css.Call("setProperty", name, value, priority)
}

func (css *CSSStyleDeclaration) Index(idx int) string {
	return css.Call("index", idx).String()
}

func (css *CSSStyleDeclaration) Length() int {
	return css.Get("length").Int()
}

type Text struct {
	*BasicNode
}
