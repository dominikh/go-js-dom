// Package dom provides GopherJS bindings for the JavaScript DOM APIs.
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
//
// Getting started
//
// The usual entry point of using the dom package is by using the
// GetWindow() function which will return a Window, from which you can
// get things such as the current Document.
//
//
// The Element, HTMLElement and Event interfaces
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
//     pEl := el.(dom.HTMLParagraphElement)
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
//     e1 := d.GetElementById("#my-element")
//     e2 := d.GetElementById("#my-element")
//
//     e1.SetClassName("some-class")
//     println(e1.ClassName() == e2.ClassName())
//
// The above example will print `true`.
package dom

import (
	"time"

	"github.com/neelance/gopherjs/js"
)

func nodeListToNodes(o js.Object) []Node {
	var out []Node
	length := o.Get("length").Int()
	for i := 0; i < length; i++ {
		out = append(out, wrapNode(o.Call("item", i)))
	}
	return out
}

func nodeListToElements(o js.Object) []Element {
	var out []Element
	length := o.Get("length").Int()
	for i := 0; i < length; i++ {
		out = append(out, wrapElement(o.Call("item", i)))
	}
	return out
}

func wrapNode(o js.Object) Node {
	if o.IsNull() || o.IsUndefined() {
		return nil
	}
	switch o.Get("constructor").Get("name").String() {
	// TODO all the non-element cases
	default:
		return wrapElement(o)
	}
}

func wrapElement(o js.Object) Element {
	if o.IsNull() || o.IsUndefined() {
		return nil
	}
	switch o.Get("constructor").Get("name").String() {
	// TODO all the non-HTML cases
	default:
		return wrapHTMLElement(o)
	}
}

func wrapHTMLElement(o js.Object) HTMLElement {
	if o.IsNull() || o.IsUndefined() {
		return nil
	}
	el := &BasicHTMLElement{&BasicElement{&BasicNode{o}}}
	name := o.Get("constructor").Get("name").String()
	switch name {
	case "HTMLAnchorElement":
		return &HTMLAnchorElement{BasicHTMLElement: el}
	case "HTMLAppletElement":
		return &HTMLAppletElement{BasicHTMLElement: el}
	case "HTMLAreaElement":
		return &HTMLAreaElement{BasicHTMLElement: el}
	case "HTMLAudioElement":
		return &HTMLAudioElement{el}
	case "HTMLBaseElement":
		return &HTMLBaseElement{el}
	case "HTMLBodyElement":
		return &HTMLBodyElement{el}
	case "HTMLBRElement":
		return &HTMLBRElement{el}
	case "HTMLButtonElement":
		return &HTMLButtonElement{BasicHTMLElement: el}
	case "HTMLCanvasElement":
		return &HTMLCanvasElement{el}
	case "HTMLDataElement":
		return &HTMLDataElement{BasicHTMLElement: el}
	case "HTMLDataListElement":
		return &HTMLDataListElement{el}
	case "HTMLDirectoryElement":
		return &HTMLDirectoryElement{el}
	case "HTMLDivElement":
		return &HTMLDivElement{el}
	case "HTMLDListElement":
		return &HTMLDListElement{el}
	case "HTMLEmbedElement":
		return &HTMLEmbedElement{BasicHTMLElement: el}
	case "HTMLFieldSetElement":
		return &HTMLFieldSetElement{el}
	case "HTMLFontElement":
		return &HTMLFontElement{el}
	case "HTMLFormElement":
		return &HTMLFormElement{el}
	case "HTMLFrameElement":
		return &HTMLFrameElement{el}
	case "HTMLFrameSetElement":
		return &HTMLFrameSetElement{el}
	case "HTMLHeadElement":
		return &HTMLHeadElement{el}
	case "HTMLHeadingElement":
		return &HTMLHeadingElement{el}
	case "HTMLHtmlElement":
		return &HTMLHtmlElement{el}
	case "HTMLHRElement":
		return &HTMLHRElement{el}
	case "HTMLIFrameElement":
		return &HTMLIFrameElement{el}
	case "HTMLImageElement":
		return &HTMLImageElement{el}
	case "HTMLInputElement":
		return &HTMLInputElement{el}
	case "HTMLKeygenElement":
		return &HTMLKeygenElement{el}
	case "HTMLLabelElement":
		return &HTMLLabelElement{BasicHTMLElement: el}
	case "HTMLLegendElement":
		return &HTMLLegendElement{el}
	case "HTMLLIElement":
		return &HTMLLIElement{BasicHTMLElement: el}
	case "HTMLLinkElement":
		return &HTMLLinkElement{el}
	case "HTMLMapElement":
		return &HTMLMapElement{el}
	case "HTMLMediaElement":
		return &HTMLMediaElement{el}
	case "HTMLMenuElement":
		return &HTMLMenuElement{el}
	case "HTMLMetaElement":
		return &HTMLMetaElement{BasicHTMLElement: el}
	case "HTMLMeterElement":
		return &HTMLMeterElement{BasicHTMLElement: el}
	case "HTMLModElement":
		return &HTMLModElement{BasicHTMLElement: el}
	case "HTMLObjectElement":
		return &HTMLObjectElement{el}
	case "HTMLOListElement":
		return &HTMLOListElement{BasicHTMLElement: el}
	case "HTMLOptGroupElement":
		return &HTMLOptGroupElement{BasicHTMLElement: el}
	case "HTMLOptionElement":
		return &HTMLOptionElement{BasicHTMLElement: el}
	case "HTMLOutputElement":
		return &HTMLOutputElement{el}
	case "HTMLParagraphElement":
		return &HTMLParagraphElement{el}
	case "HTMLParamElement":
		return &HTMLParamElement{BasicHTMLElement: el}
	case "HTMLPreElement":
		return &HTMLPreElement{el}
	case "HTMLProgressElement":
		return &HTMLProgressElement{BasicHTMLElement: el}
	case "HTMLQuoteElement":
		return &HTMLQuoteElement{BasicHTMLElement: el}
	case "HTMLScriptElement":
		return &HTMLScriptElement{el}
	case "HTMLSelectElement":
		return &HTMLSelectElement{el}
	case "HTMLSourceElement":
		return &HTMLSourceElement{BasicHTMLElement: el}
	case "HTMLSpanElement":
		return &HTMLSpanElement{el}
	case "HTMLStyleElement":
		return &HTMLStyleElement{el}
	case "HTMLTableElement":
		return &HTMLTableElement{el}
	case "HTMLTableCaptionElement":
		return &HTMLTableCaptionElement{el}
	case "HTMLTableCellElement":
		return &HTMLTableCellElement{el}
	case "HTMLTableDataCellElement":
		return &HTMLTableDataCellElement{el}
	case "HTMLTableHeaderCellElement":
		return &HTMLTableHeaderCellElement{BasicHTMLElement: el}
	case "HTMLTableColElement":
		return &HTMLTableColElement{BasicHTMLElement: el}
	case "HTMLTableRowElement":
		return &HTMLTableRowElement{el}
	case "HTMLTableSectionElement":
		return &HTMLTableSectionElement{el}
	case "HTMLTextAreaElement":
		return &HTMLTextAreaElement{el}
	case "HTMLTimeElement":
		return &HTMLTimeElement{BasicHTMLElement: el}
	case "HTMLTitleElement":
		return &HTMLTitleElement{BasicHTMLElement: el}
	case "HTMLTrackElement":
		return &HTMLTrackElement{el}
	case "HTMLUListElement":
		return &HTMLUListElement{el}
	case "HTMLUnknownElement":
		return &HTMLUnknownElement{el}
	case "HTMLVideoElement":
		return &HTMLVideoElement{el}
	default:
		panic("Unsupported HTML element type: " + name)
	}
}

func wrapEvent(o js.Object) Event {
	if o.IsNull() || o.IsUndefined() {
		return nil
	}
	ev := &BasicEvent{o}
	name := o.Get("constructor").Get("name").String()
	switch name {
	case "AnimationEvent":
		return &AnimationEvent{ev}
	case "AudioProcessingEvent":
		return &AudioProcessingEvent{ev}
	case "BeforeInputEvent":
		return &BeforeInputEvent{ev}
	case "BeforeUnloadEvent":
		return &BeforeUnloadEvent{ev}
	case "BlobEvent":
		return &BlobEvent{ev}
	case "ClipboardEvent":
		return &ClipboardEvent{ev}
	case "CloseEvent":
		return &CloseEvent{ev}
	case "CompositionEvent":
		return &CompositionEvent{ev}
	case "CSSFontFaceLoadEvent":
		return &CSSFontFaceLoadEvent{ev}
	case "CustomEvent":
		return &CustomEvent{ev}
	case "DeviceLightEvent":
		return &DeviceLightEvent{ev}
	case "DeviceMotionEvent":
		return &DeviceMotionEvent{ev}
	case "DeviceOrientationEvent":
		return &DeviceOrientationEvent{ev}
	case "DeviceProximityEvent":
		return &DeviceProximityEvent{ev}
	case "DOMTransactionEvent":
		return &DOMTransactionEvent{ev}
	case "DragEvent":
		return &DragEvent{ev}
	case "EditingBeforeInputEvent":
		return &EditingBeforeInputEvent{ev}
	case "ErrorEvent":
		return &ErrorEvent{ev}
	case "FocusEvent":
		return &FocusEvent{ev}
	case "GamepadEvent":
		return &GamepadEvent{ev}
	case "HashChangeEvent":
		return &HashChangeEvent{ev}
	case "IDBVersionChangeEvent":
		return &IDBVersionChangeEvent{ev}
	case "KeyboardEvent":
		return &KeyboardEvent{ev}
	case "MediaStreamEvent":
		return &MediaStreamEvent{ev}
	case "MessageEvent":
		return &MessageEvent{ev}
	case "MouseEvent":
		return &MouseEvent{ev}
	case "MutationEvent":
		return &MutationEvent{ev}
	case "OfflineAudioCompletionEvent":
		return &OfflineAudioCompletionEvent{ev}
	case "PageTransitionEvent":
		return &PageTransitionEvent{ev}
	case "PointerEvent":
		return &PointerEvent{ev}
	case "PopStateEvent":
		return &PopStateEvent{ev}
	case "ProgressEvent":
		return &ProgressEvent{ev}
	case "RelatedEvent":
		return &RelatedEvent{ev}
	case "RTCPeerConnectionIceEvent":
		return &RTCPeerConnectionIceEvent{ev}
	case "SensorEvent":
		return &SensorEvent{ev}
	case "StorageEvent":
		return &StorageEvent{ev}
	case "SVGEvent":
		return &SVGEvent{ev}
	case "SVGZoomEvent":
		return &SVGZoomEvent{ev}
	case "TimeEvent":
		return &TimeEvent{ev}
	case "TouchEvent":
		return &TouchEvent{ev}
	case "TrackEvent":
		return &TrackEvent{ev}
	case "TransitionEvent":
		return &TransitionEvent{ev}
	case "UIEvent":
		return &UIEvent{ev}
	case "UserProximityEvent":
		return &UserProximityEvent{ev}
	case "WheelEvent":
		return &WheelEvent{ev}
	default:
		return ev
	}
}

func GetWindow() Window {
	return &window{js.Global("window")}
}

// TODO document that our lists aren't live

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
	CreateElement(name string) Element
	CreateElementNS(namespace, name string) Element
	ElementFromPoint(x, y int) Element
	EnableStyleSheetsForSet(name string)
	GetElementsByClassName(name string) []Element
	GetElementsByTagName(name string) []Element
	GetElementsByTagNameNS(ns, name string) []Element
	GetElementByID(id string) Element
	QuerySelector(sel string) Element
	QuerySelectorAll(sel string) []Element
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
	Location() Location
	Plugins() []*HTMLEmbedElement
	ReadyState() string
	Referrer() string
	Scripts() []*HTMLScriptElement
	Title() string
	SetTitle(string)
	URL() string

	// TODO HTMLDocument methods
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
	if s == "off" {
		return false
	}
	return true
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
	return d.Get("lastModified").Interface().(time.Time)
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

func (d *htmlDocument) Location() Location {
	return Location{d.Get("location")}
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
	return d.Get("url").String()
}

func (d document) Async() bool {
	return d.Get("async").Bool()
}

func (d document) SetAsync(b bool) {
	d.Set("async", b)
}

func (d document) Doctype() DocumentType {
	// FIXME implement
	return nil
}

func (d document) DocumentElement() Element {
	return wrapElement(d.Get("documentElement"))
}

func (d document) DocumentURI() string {
	return d.Get("documentURI").String()
}

func (d document) Implementation() DOMImplementation {
	// FIXME implement
	return nil
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
	return nil
}

func (d document) StyleSheetSets() []StyleSheet {
	// FIXME implement
	return nil
}

func (d document) AdoptNode(node Node) Node {
	return wrapNode(d.Call("adoptNode", node.Underlying()))
}

func (d document) CreateElement(name string) Element {
	return wrapElement(d.Call("createElement", name))
}

func (d document) CreateElementNS(ns string, name string) Element {
	return wrapElement(d.Call("createElement", ns, name))
}

func (d document) ElementFromPoint(x, y int) Element {
	return wrapElement(d.Call("elementFromPoint", x, y))
}

func (d document) EnableStyleSheetsForSet(name string) {
	d.Call("enableStyleSheetsForSet", name)
}

func (d document) GetElementsByClassName(name string) []Element {
	return (&BasicElement{&BasicNode{d.Object}}).GetElementsByClassName(name)
}

func (d document) GetElementsByTagName(name string) []Element {
	return (&BasicElement{&BasicNode{d.Object}}).GetElementsByTagName(name)
}

func (d document) GetElementsByTagNameNS(ns, name string) []Element {
	return (&BasicElement{&BasicNode{d.Object}}).GetElementsByTagNameNS(ns, name)
}

func (d document) GetElementByID(id string) Element {
	return wrapElement(d.Call("getElementById", id))
}

func (d document) QuerySelector(sel string) Element {
	return (&BasicElement{&BasicNode{d.Object}}).QuerySelector(sel)
}

func (d document) QuerySelectorAll(sel string) []Element {
	return (&BasicElement{&BasicNode{d.Object}}).QuerySelectorAll(sel)
}

type Location struct{ js.Object }

// TODO Location methods

type HTMLElement interface {
	Element
	GlobalEventHandlers

	AccessKey() string
	SetAccessKey(string)
	AccessKeyLabel() string
	SetAccessKeyLabel(string)
	ContentEditable() string
	SetContentEditable(string)
	IsContentEditable() bool
	Dataset() // FIXME type
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
	ClearInterval(int)
	ClearTimeout(int)
	Close()
	Confirm(string) bool
	Focus()
	Forward()
	GetComputedStyle(el Element) *CSSStyleDeclaration
	GetSelection() Selection
	Home()
	MoveBy(dx, dy int)
	MoveTo(x, y int)
	Open(url, name, features string) Window
	OpenDialog(url, name, features string, args []interface{}) Window
	PostMessage(message string, target string, transfer []interface{})
	Print()
	Prompt(prompt string, initial string) string
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
	js.Object
}

func (w *window) Console() *Console {
	return &Console{w.Get("console")}
}

func (w *window) Document() Document {
	// TODO return a HTMLDocument if possible
	o := w.Get("document")
	switch o.Get("constructor").Get("name").String() {
	case "HTMLDocument":
		return &htmlDocument{&document{&BasicNode{w.Get("document")}}}
	default:
		return &document{&BasicNode{w.Get("document")}}
	}
}

func (w *window) FrameElement() Element {
	return wrapElement(w.Get("frameElement"))
}

func (w *window) Location() *Location {
	return &Location{w.Get("location")}
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
	return nil
}

func (w *window) Screen() *Screen {
	return &Screen{Object: w.Get("screen")}
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

func (w *window) GetComputedStyle(el Element) *CSSStyleDeclaration {
	// FIXME implement
	return &CSSStyleDeclaration{}
}

func (w *window) GetSelection() Selection {
	// FIXME implement
	return nil
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
	return w.Call("setInterval", fn, delay).Int()
}

func (w *window) SetTimeout(fn func(), delay int) int {
	return w.Call("setTimeout", fn, delay).Int()
}

func (w *window) Stop() {
	w.Call("stop")
}

func (w *window) AddEventListener(typ string, useCapture bool, listener func(Event)) {
	w.Call("addEventListener", typ, func(o js.Object) { listener(wrapEvent(o)) }, useCapture)
}
func (w *window) RemoveEventListener(typ string, useCapture bool, listener func(Event)) {
	w.Call("removeEventListener", typ, func(o js.Object) { listener(wrapEvent(o)) }, useCapture)
}

// TODO all the other window methods

type Selection interface {
	// TODO
}

type Screen struct {
	js.Object
	AvailTop    int `js:"availTop"`
	AvailLeft   int `js:"availLeft"`
	AvailHeight int `js:"availHeight"`
	AvailWidth  int `js:"availWidth"`
	ColorDepth  int `js:"colorDepth"`
	Height      int `js:"height"`
	Left        int `js:"left"`
	PixelDepth  int `js:"pixelDepth"`
	Top         int `js:"top"`
	Width       int `js:"width"`
}

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
	js.Object
	Code int `js:code"`
}

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
	js.Object
	Latitude         float64 `js:"latitude"`
	Longitude        float64 `js:"longitude"`
	Altitude         float64 `js:"altitude"`
	Accuracy         float64 `js:"accuracy"`
	AltitudeAccuracy float64 `js:"altitudeAccuracy"`
	Heading          float64 `js:"heading"`
	Speed            float64 `js:"speed"`
}

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
	js.Object
	// TODO will replace the js/console package
}

type SVGDocument interface{}
type DocumentType interface{}
type DOMImplementation interface{}
type StyleSheet interface{}
type CSSStyleSheet interface{}

type EventTarget interface {
	AddEventListener(typ string, useCapture bool, listener func(Event))
	RemoveEventListener(typ string, useCapture bool, listener func(Event))
	// DispatchEvent() // TODO
}

type Node interface {
	// Due to issue 7158, we cannot embed EventTarget right now

	// EventTarget // FIXME
	AddEventListener(typ string, useCapture bool, listener func(Event))    // FIXME
	RemoveEventListener(typ string, useCapture bool, listener func(Event)) // FIXME

	Underlying() js.Object
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
	AppendChild(Element)
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
	js.Object
}

func (n *BasicNode) Underlying() js.Object {
	return n.Object
}

func (n *BasicNode) AddEventListener(typ string, useCapture bool, listener func(Event)) {
	n.Call("addEventListener", typ, func(o js.Object) { listener(wrapEvent(o)) }, useCapture)
}

func (n *BasicNode) RemoveEventListener(typ string, useCapture bool, listener func(Event)) {
	n.Call("removeEventListener", typ, func(o js.Object) { listener(wrapEvent(o)) }, useCapture)
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
	return n.Get("nodeValue").String()
}

func (n *BasicNode) SetNodeValue(s string) {
	n.Set("nodeValue", s)
}

func (n *BasicNode) OwnerDocument() Document {
	// FIXME implement
	return nil
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
	return n.Get("textContent").String()
}

func (n *BasicNode) SetTextContent(s string) {
	n.Set("textContent", s)
}

func (n *BasicNode) AppendChild(e Element) {
	n.Call("appendChild", e.Underlying())
}

func (n *BasicNode) CloneNode(deep bool) Node {
	return wrapNode(n.Call("cloneNode", deep))
}

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
	n.Call("insertBefore", which.Underlying(), before.Underlying())
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
	return n.Call("lookupNamespaceURI", s).String()
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

	// We do not have classList and no DOMTokenList. manipulate
	// strings yourself
	// TODO ClassName → Class?
	ClassName() string
	SetClassName(string)
	ID() string
	SetID(string)
	TagName() string
	GetAttribute(string) string                   // TODO can attributes only be strings?
	GetAttributeNS(ns string, name string) string // can attributes only be strings?
	GetElementsByClassName(string) []Element
	GetElementsByTagName(string) []Element
	GetElementsByTagNameNS(ns string, name string) []Element
	HasAttribute(string) bool
	HasAttributeNS(ns string, name string) bool
	QuerySelector(string) Element
	QuerySelectorAll(string) []Element
	RemoveAttribute(string)
	RemoveAttributeNS(ns string, name string)
	SetAttribute(name string, value string)
	SetAttributeNS(ns string, name string, value string)
	AddClass(string)
	RemoveClass(string)
	ToggleClass(string)
	HasClass(string)
	InnerHTML() string
	SetInnerHTML(string)
}

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

func (e *BasicHTMLElement) Dataset() {
	// FIXME type
	// FIXME implement
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

func (e *BasicElement) PreviousElementSibling() Element {
	return wrapElement(e.Get("previousElementSibling"))
}

func (e *BasicElement) NextElementSibling() Element {
	return wrapElement(e.Get("nextElementSibling"))
}

func (e *BasicElement) ClassName() string {
	return e.Get("className").String()
}

func (e *BasicElement) SetClassName(s string) {
	e.Set("className", s)
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
	return e.Call("getAttribute", name).String()
}

func (e *BasicElement) GetAttributeNS(ns string, name string) string {
	return e.Call("getAttributeNS", ns, name).String()
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

func (e *BasicElement) AddClass(string) {
	// FIXME implement
}

func (e *BasicElement) RemoveClass(string) {
	// FIXME implement
}

func (e *BasicElement) ToggleClass(string) {
	// FIXME implement
}

func (e *BasicElement) HasClass(string) {
	// FIXME implement
}

func (e *BasicElement) InnerHTML() string {
	return e.Get("innerHTML").String()
}

func (e *BasicElement) SetInnerHTML(s string) {
	e.Set("innerHTML", s)
}

type HTMLAnchorElement struct {
	*BasicHTMLElement
	HrefLang string `js:"hreflang"`
	Media    string `js:"media"`
	Rel      string `js:"rel"`
	TabIndex int    `js:"tabIndex"`
	Target   string `js:"target"`
	Text     string `js:"text"`
	Type     string `js:"type"`
}

type HTMLAppletElement struct {
	*BasicHTMLElement
	Alt      string `js:"alt"`
	Coords   string `js:"coords"`
	HrefLang string `js:"hreflang"`
	Media    string `js:"media"`
	Rel      string `js:"rel"`
	Search   string `js:"search"`
	Shape    string `js:"shape"`
	TabIndex int    `js:"tabIndex"`
	Target   string `js:"target"`
	Type     string `js:"type"`
}

type HTMLAreaElement struct {
	*BasicHTMLElement
	Alt      string `js:"alt"`
	Coords   string `js:"coords"`
	HrefLang string `js:"hreflang"`
	Media    string `js:"media"`
	Rel      string `js:"rel"`
	Search   string `js:"search"`
	Shape    string `js:"shape"`
	TabIndex int    `js:"tabIndex"`
	Target   string `js:"target"`
	Type     string `js:"type"`
}

type HTMLAudioElement struct{ *BasicHTMLElement }

type HTMLBRElement struct{ *BasicHTMLElement }
type HTMLBaseElement struct{ *BasicHTMLElement }
type HTMLBodyElement struct{ *BasicHTMLElement }

type ValidityState struct {
	js.Object
	CustomError     bool `js:"customError"`
	PatternMismatch bool `js:"	patternMismatch"`
	RangeOverflow   bool `js:"	rangeOverflow"`
	RangeUnderflow  bool `js:"	rangeUnderflow"`
	StepMismatch    bool `js:"	stepMismatch"`
	TooLong         bool `js:"	tooLong"`
	TypeMismatch    bool `js:"	typeMismatch"`
	Valid           bool `js:"	valid"`
	ValueMissing    bool `js:"	valueMissing"`
}

type HTMLButtonElement struct {
	*BasicHTMLElement
	AutoFocus         bool   `js:"autofocus"`
	Disabled          bool   `js:"disabled"`
	FormAction        string `js:"formAction"`
	FormEncType       string `js:"formEncType"`
	FormMethod        string `js:"formMethod"`
	FormNoValidate    bool   `js:"formNoValidate"`
	FormTarget        string `js:"formTarget"`
	Name              string `js:"name"`
	TabIndex          int    `js:"tabIndex"`
	Type              string `js:"type"`
	ValidationMessage string `js:"validationMessage"`
	Value             string `js:"value"`
	WillValidate      bool   `js:"willValidate"`
}

func (e *HTMLButtonElement) Form() *HTMLFormElement {
	form := wrapHTMLElement(e.Get("form"))
	if form == nil {
		return nil
	}
	return form.(*HTMLFormElement)
}

func (e *HTMLButtonElement) Labels() []*HTMLLabelElement {
	labels := nodeListToElements(e.Get("labels"))
	out := make([]*HTMLLabelElement, len(labels))
	for i, label := range labels {
		out[i] = label.(*HTMLLabelElement)
	}
	return out
}

func (e *HTMLButtonElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Object: e.Get("validity")}
}

type HTMLCanvasElement struct{ *BasicHTMLElement }
type HTMLDListElement struct{ *BasicHTMLElement }

type HTMLDataElement struct {
	*BasicHTMLElement
	Value string `js:"value"`
}

type HTMLDataListElement struct{ *BasicHTMLElement }
type HTMLDirectoryElement struct{ *BasicHTMLElement }
type HTMLDivElement struct{ *BasicHTMLElement }

type HTMLEmbedElement struct {
	*BasicHTMLElement
	Src   string `js:"src"`
	Type  string `js:"type"`
	Width string `js:"width"`
}

type HTMLFieldSetElement struct{ *BasicHTMLElement }
type HTMLFontElement struct{ *BasicHTMLElement }
type HTMLFormElement struct{ *BasicHTMLElement }
type HTMLFrameElement struct{ *BasicHTMLElement }
type HTMLFrameSetElement struct{ *BasicHTMLElement }
type HTMLHRElement struct{ *BasicHTMLElement }
type HTMLHeadElement struct{ *BasicHTMLElement }
type HTMLHeadingElement struct{ *BasicHTMLElement }
type HTMLHtmlElement struct{ *BasicHTMLElement }
type HTMLIFrameElement struct{ *BasicHTMLElement }
type HTMLImageElement struct{ *BasicHTMLElement }
type HTMLInputElement struct{ *BasicHTMLElement }
type HTMLKeygenElement struct{ *BasicHTMLElement }

type HTMLLIElement struct {
	*BasicHTMLElement
	Value int `js:"value"`
}

type HTMLLabelElement struct {
	*BasicHTMLElement
	For string `js:"htmlFor"`
}

func (e *HTMLLabelElement) Control() HTMLElement {
	return wrapHTMLElement(e.Get("control"))
}

func (e *HTMLLabelElement) Form() *HTMLFormElement {
	form := wrapHTMLElement(e.Get("form"))
	if form == nil {
		return nil
	}
	return form.(*HTMLFormElement)
}

type HTMLLegendElement struct{ *BasicHTMLElement }

func (e *HTMLLegendElement) Form() *HTMLFormElement {
	form := wrapHTMLElement(e.Get("form"))
	if form == nil {
		return nil
	}
	return form.(*HTMLFormElement)
}

type HTMLLinkElement struct{ *BasicHTMLElement }
type HTMLMapElement struct{ *BasicHTMLElement }
type HTMLMediaElement struct{ *BasicHTMLElement }
type HTMLMenuElement struct{ *BasicHTMLElement }

type HTMLMetaElement struct {
	*BasicHTMLElement
	Content   string `js:"content"`
	HTTPEquiv string `js:"httpEquiv"`
	Name      string `js:"name"`
}

type HTMLMeterElement struct {
	*BasicHTMLElement
	High    float64 `js:"high"`
	Low     float64 `js:"low"`
	Max     float64 `js:"max"`
	Min     float64 `js:"min"`
	Optimum float64 `js:"optimum"`
}

func (e HTMLMeterElement) Labels() []*HTMLLabelElement {
	nodes := nodeListToNodes(e.Call("labels"))
	out := make([]*HTMLLabelElement, len(nodes))
	for i, node := range nodes {
		out[i] = node.(*HTMLLabelElement)
	}
	return out
}

type HTMLModElement struct {
	*BasicHTMLElement
	Cite     string `js:"cite"`
	DateTime string `js:"dateTime"`
}

type HTMLOListElement struct {
	*BasicHTMLElement
	Reversed bool   `js:"reversed"`
	Start    int    `js:"start"`
	Type     string `js:"type"`
}

type HTMLObjectElement struct{ *BasicHTMLElement }

type HTMLOptGroupElement struct {
	*BasicHTMLElement
	Disabled bool   `js:"disabled"`
	Label    string `js:"label"`
}

type HTMLOptionElement struct {
	*BasicHTMLElement
	DefaultSelected bool   `js:"defaultSelected"`
	Disabled        bool   `js:"disabled"`
	Index           int    `js:"index"`
	Label           string `js:"label"`
	Selected        bool   `js:"selected"`
	Text            string `js:"text"`
	Value           string `js:"value"`
}

func (e *HTMLOptionElement) Form() *HTMLFormElement {
	el := wrapHTMLElement(e.Get("form"))
	if el == nil {
		return nil
	}
	return el.(*HTMLFormElement)
}

type HTMLOutputElement struct{ *BasicHTMLElement }
type HTMLParagraphElement struct{ *BasicHTMLElement }

type HTMLParamElement struct {
	*BasicHTMLElement
	Name  string `js:"name"`
	Value string `js:"value"`
}

type HTMLPreElement struct{ *BasicHTMLElement }

type HTMLProgressElement struct {
	*BasicHTMLElement
	Max      float64 `js:"max"`
	Position float64 `js:"position"`
	Value    float64 `js:"value"`
}

func (e HTMLProgressElement) Labels() []*HTMLLabelElement {
	nodes := nodeListToNodes(e.Call("labels"))
	out := make([]*HTMLLabelElement, len(nodes))
	for i, node := range nodes {
		out[i] = node.(*HTMLLabelElement)
	}
	return out
}

type HTMLQuoteElement struct {
	*BasicHTMLElement
	Cite string `js:"cite"`
}

type HTMLScriptElement struct{ *BasicHTMLElement }
type HTMLSelectElement struct{ *BasicHTMLElement }

type HTMLSourceElement struct {
	*BasicHTMLElement
	Media string `js:"media"`
	Src   string `js:"src"`
	Type  string `js:"type"`
}

type HTMLSpanElement struct{ *BasicHTMLElement }
type HTMLStyleElement struct{ *BasicHTMLElement }
type HTMLTableCaptionElement struct{ *BasicHTMLElement }
type HTMLTableCellElement struct{ *BasicHTMLElement }

type HTMLTableColElement struct {
	*BasicHTMLElement
	Span int `js:"span"`
}

type HTMLTableDataCellElement struct{ *BasicHTMLElement }
type HTMLTableElement struct{ *BasicHTMLElement }

type HTMLTableHeaderCellElement struct {
	*BasicHTMLElement
	Abbr  string `js:"abbr"`
	Scope string `js:"scope"`
}

type HTMLTableRowElement struct{ *BasicHTMLElement }
type HTMLTableSectionElement struct{ *BasicHTMLElement }
type HTMLTextAreaElement struct{ *BasicHTMLElement }

type HTMLTimeElement struct {
	*BasicHTMLElement
	DateTime string `js:"dateTime"`
}

type HTMLTitleElement struct {
	*BasicHTMLElement
	Text string `js:"text"`
}

type HTMLTrackElement struct{ *BasicHTMLElement }
type HTMLUListElement struct{ *BasicHTMLElement }
type HTMLUnknownElement struct{ *BasicHTMLElement }
type HTMLVideoElement struct{ *BasicHTMLElement }

func (e *HTMLBaseElement) Href() string {
	return e.Get("href").String()
}

func (e *HTMLBaseElement) Target() string {
	return e.Get("target").String()
}

const (
	EvPhaseNone      = 0
	EvPhaseCapturing = 1
	EvPhaseAtTarget  = 2
	EvPhaseBubbling  = 3
)

type Event interface {
	Bubbles() bool
	Cancelable() bool
	CurrentTarget() Element
	DefaultPrevented() bool
	EventPhase() int
	Target() Element
	Timestamp() time.Time
	Type() string
	PreventDefault()
	StopImmediatePropagation()
	StopPropagation()
}

// Type BasicEvent implements the Event interface and is embedded by
// concrete event types.
type BasicEvent struct{ js.Object }

func (ev *BasicEvent) Bubbles() bool {
	return ev.Get("bubbles").Bool()
}

func (ev *BasicEvent) Cancelable() bool {
	return ev.Get("cancelable").Bool()
}

func (ev *BasicEvent) CurrentTarget() Element {
	return wrapElement(ev.Get("currentTarget"))
}

func (ev *BasicEvent) DefaultPrevented() bool {
	return ev.Get("defaultPrevented").Bool()
}

func (ev *BasicEvent) EventPhase() int {
	return ev.Get("eventPhase").Int()
}

func (ev *BasicEvent) Target() Element {
	return wrapElement(ev.Get("target"))
}

func (ev *BasicEvent) Timestamp() time.Time {
	ms := ev.Get("timeStamp").Int()
	s := ms / 1000
	ns := (ms % 1000 * 1e6)
	return time.Unix(int64(s), int64(ns))
}

func (ev *BasicEvent) Type() string {
	return ev.Get("type").String()
}

func (ev *BasicEvent) PreventDefault() {
	ev.Call("preventDefault")
}

func (ev *BasicEvent) StopImmediatePropagation() {
	ev.Call("stopImmediatePropagation")
}

func (ev *BasicEvent) StopPropagation() {
	ev.Call("stopPropagation")
}

type AnimationEvent struct{ *BasicEvent }
type AudioProcessingEvent struct{ *BasicEvent }
type BeforeInputEvent struct{ *BasicEvent }
type BeforeUnloadEvent struct{ *BasicEvent }
type BlobEvent struct{ *BasicEvent }
type ClipboardEvent struct{ *BasicEvent }
type CloseEvent struct{ *BasicEvent }
type CompositionEvent struct{ *BasicEvent }
type CSSFontFaceLoadEvent struct{ *BasicEvent }
type CustomEvent struct{ *BasicEvent }
type DeviceLightEvent struct{ *BasicEvent }
type DeviceMotionEvent struct{ *BasicEvent }
type DeviceOrientationEvent struct{ *BasicEvent }
type DeviceProximityEvent struct{ *BasicEvent }
type DOMTransactionEvent struct{ *BasicEvent }
type DragEvent struct{ *BasicEvent }
type EditingBeforeInputEvent struct{ *BasicEvent }
type ErrorEvent struct{ *BasicEvent }
type FocusEvent struct{ *BasicEvent }
type GamepadEvent struct{ *BasicEvent }
type HashChangeEvent struct{ *BasicEvent }
type IDBVersionChangeEvent struct{ *BasicEvent }
type KeyboardEvent struct{ *BasicEvent }
type MediaStreamEvent struct{ *BasicEvent }
type MessageEvent struct{ *BasicEvent }
type MouseEvent struct{ *BasicEvent }
type MutationEvent struct{ *BasicEvent }
type OfflineAudioCompletionEvent struct{ *BasicEvent }
type PageTransitionEvent struct{ *BasicEvent }
type PointerEvent struct{ *BasicEvent }
type PopStateEvent struct{ *BasicEvent }
type ProgressEvent struct{ *BasicEvent }
type RelatedEvent struct{ *BasicEvent }
type RTCPeerConnectionIceEvent struct{ *BasicEvent }
type SensorEvent struct{ *BasicEvent }
type StorageEvent struct{ *BasicEvent }
type SVGEvent struct{ *BasicEvent }
type SVGZoomEvent struct{ *BasicEvent }
type TimeEvent struct{ *BasicEvent }
type TouchEvent struct{ *BasicEvent }
type TrackEvent struct{ *BasicEvent }
type TransitionEvent struct{ *BasicEvent }
type UIEvent struct{ *BasicEvent }
type UserProximityEvent struct{ *BasicEvent }
type WheelEvent struct{ *BasicEvent }
type CSSStyleDeclaration struct{ js.Object }

func (css *CSSStyleDeclaration) ToMap() map[string]string {
	m := make(map[string]string)
	N := css.Get("length").Int()
	for i := 0; i < N; i++ {
		name := css.Call("index", i).String()
		value := css.Call("getPropertyValue").String()
		m[name] = value
	}

	return m
}

func (css *CSSStyleDeclaration) RemoveProperty(name string) {
	css.Call("remoteProperty", name)
}

func (css *CSSStyleDeclaration) GetPropertyValue(name string) string {
	return css.Call("getPropertyValue", name).String()
}

func (css *CSSStyleDeclaration) GetPropertyPriority(name string) string {
	return css.Call("getPropertyPriority", name).String()
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
