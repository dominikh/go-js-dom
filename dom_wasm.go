// +build js,wasm

package dom

import (
	"image"
	"image/color"
	"strings"
	"syscall/js"
	"time"
)

// Keys returns the keys of the given JavaScript object.
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

	Length int `js:"length"`
}

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
	return wrapElement(d.Object.Call("getElementById", id))
}

func (d documentFragment) QuerySelector(sel string) Element {
	return (&BasicElement{&BasicNode{d.Object}}).QuerySelector(sel)
}

func (d documentFragment) QuerySelectorAll(sel string) []Element {
	return (&BasicElement{&BasicNode{d.Object}}).QuerySelectorAll(sel)
}

type document struct {
	*BasicNode
}

type htmlDocument struct {
	*document
}

func (d *htmlDocument) ActiveElement() HTMLElement {
	return wrapHTMLElement(d.Object.Get("activeElement"))
}

func (d *htmlDocument) Body() HTMLElement {
	return wrapHTMLElement(d.Object.Get("body"))
}

func (d *htmlDocument) Cookie() string {
	return d.Object.Get("cookie").String()
}

func (d *htmlDocument) SetCookie(s string) {
	d.Object.Set("cookie", s)
}

func (d *htmlDocument) DefaultView() Window {
	return &window{d.Object.Get("defaultView")}
}

func (d *htmlDocument) DesignMode() bool {
	s := d.Object.Get("designMode").String()
	return s != "off"
}

func (d *htmlDocument) SetDesignMode(b bool) {
	s := "off"
	if b {
		s = "on"
	}
	d.Object.Set("designMode", s)
}

func (d *htmlDocument) Domain() string {
	return d.Object.Get("domain").String()
}

func (d *htmlDocument) SetDomain(s string) {
	d.Object.Set("domain", s)
}

func (d *htmlDocument) Forms() []*HTMLFormElement {
	var els []*HTMLFormElement
	forms := d.Object.Get("forms")
	length := forms.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(forms.Call("item", i)).(*HTMLFormElement))
	}
	return els
}

func (d *htmlDocument) Head() *HTMLHeadElement {
	head := wrapElement(d.Object.Get("head"))
	if head == nil {
		return nil
	}
	return head.(*HTMLHeadElement)
}

func (d *htmlDocument) Images() []*HTMLImageElement {
	var els []*HTMLImageElement
	images := d.Object.Get("images")
	length := images.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(images.Call("item", i)).(*HTMLImageElement))
	}
	return els
}

func (d *htmlDocument) LastModified() time.Time {
	return time.Unix(0, int64(d.Object.Get("lastModified").Call("getTime").Int())*1000000)
}

func (d *htmlDocument) Links() []HTMLElement {
	var els []HTMLElement
	links := d.Object.Get("links")
	length := links.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(links.Call("item", i)))
	}
	return els
}

func (d *htmlDocument) Location() *Location {
	o := d.Object.Get("location")
	return &Location{Value: o, URLUtils: &URLUtils{Value: o}}
}

func (d *htmlDocument) Plugins() []*HTMLEmbedElement {
	var els []*HTMLEmbedElement
	forms := d.Object.Get("plugins")
	length := forms.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(forms.Call("item", i)).(*HTMLEmbedElement))
	}
	return els
}

func (d *htmlDocument) ReadyState() string {
	return d.Object.Get("readyState").String()
}

func (d *htmlDocument) Referrer() string {
	return d.Object.Get("referrer").String()
}

func (d *htmlDocument) Scripts() []*HTMLScriptElement {
	var els []*HTMLScriptElement
	forms := d.Object.Get("scripts")
	length := forms.Get("length").Int()
	for i := 0; i < length; i++ {
		els = append(els, wrapHTMLElement(forms.Call("item", i)).(*HTMLScriptElement))
	}
	return els
}

func (d *htmlDocument) Title() string {
	return d.Object.Get("title").String()
}

func (d *htmlDocument) SetTitle(s string) {
	d.Object.Set("title", s)
}

func (d *htmlDocument) URL() string {
	return d.Object.Get("URL").String()
}

func (d document) Async() bool {
	return d.Object.Get("async").Bool()
}

func (d document) SetAsync(b bool) {
	d.Object.Set("async", b)
}

func (d document) Doctype() DocumentType {
	// FIXME implement
	panic("not implemented")
}

func (d document) DocumentElement() Element {
	return wrapElement(d.Object.Get("documentElement"))
}

func (d document) DocumentURI() string {
	return d.Object.Get("documentURI").String()
}

func (d document) Implementation() DOMImplementation {
	// FIXME implement
	panic("not implemented")
}

func (d document) LastStyleSheetSet() string {
	return d.Object.Get("lastStyleSheetSet").String()
}

func (d document) PreferredStyleSheetSet() string {
	return d.Object.Get("preferredStyleSheetSet").String()
}

func (d document) SelectedStyleSheetSet() string {
	return d.Object.Get("selectedStyleSheetSet").String()
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
	return wrapNode(d.Object.Call("adoptNode", node.Underlying()))
}

func (d document) ImportNode(node Node, deep bool) Node {
	return wrapNode(d.Object.Call("importNode", node.Underlying(), deep))
}

func (d document) CreateDocumentFragment() DocumentFragment {
	return wrapDocumentFragment(d.Object.Call("createDocumentFragment"))
}

func (d document) CreateElement(name string) Element {
	return wrapElement(d.Object.Call("createElement", name))
}

func (d document) CreateElementNS(ns string, name string) Element {
	return wrapElement(d.Object.Call("createElementNS", ns, name))
}

func (d document) CreateTextNode(s string) *Text {
	return wrapNode(d.Object.Call("createTextNode", s)).(*Text)
}

func (d document) ElementFromPoint(x, y int) Element {
	return wrapElement(d.Object.Call("elementFromPoint", x, y))
}

func (d document) EnableStyleSheetsForSet(name string) {
	d.Object.Call("enableStyleSheetsForSet", name)
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
	return wrapElement(d.Object.Call("getElementById", id))
}

func (d document) QuerySelector(sel string) Element {
	return (&BasicElement{&BasicNode{d.Object}}).QuerySelector(sel)
}

func (d document) QuerySelectorAll(sel string) []Element {
	return (&BasicElement{&BasicNode{d.Object}}).QuerySelectorAll(sel)
}

type URLUtils struct {
	js.Value

	Href     string `js:"href"`
	Protocol string `js:"protocol"`
	Host     string `js:"host"`
	Hostname string `js:"hostname"`
	Port     string `js:"port"`
	Pathname string `js:"pathname"`
	Search   string `js:"search"`
	Hash     string `js:"hash"`
	Username string `js:"username"`
	Password string `js:"password"`
	Origin   string `js:"origin"`
}

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
	return w.Call("setInterval", fn, delay).Int()
}

func (w *window) SetTimeout(fn func(), delay int) int {
	return w.Call("setTimeout", fn, delay).Int()
}

func (w *window) Stop() {
	w.Call("stop")
}

// TODO reuse util.EventTarget

func (w *window) AddEventListener(typ string, useCapture bool, listener func(Event)) func(o js.Value) {
	wrapper := func(o js.Value) { listener(wrapEvent(o)) }
	w.Call("addEventListener", typ, wrapper, useCapture)
	return wrapper
}

func (w *window) RemoveEventListener(typ string, useCapture bool, listener func(js.Value)) {
	w.Call("removeEventListener", typ, listener, useCapture)
}

func (w *window) DispatchEvent(event Event) bool {
	return w.Call("dispatchEvent", event).Bool()
}

func wrapDOMHighResTimeStamp(o js.Value) time.Duration {
	return time.Duration(o.Float() * float64(time.Millisecond))
}

func (w *window) RequestAnimationFrame(callback func(time.Duration)) int {
	wrapper := func(o js.Value) { callback(wrapDOMHighResTimeStamp(o)) }
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
	js.Value
	Code int `js:"code"`
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
	js.Value
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
	// Do not use Value because most of the struct that embed this one already has Value of type string.
	Object js.Value
}

func (n *BasicNode) Underlying() js.Value {
	return n.Object
}

func (n *BasicNode) AddEventListener(typ string, useCapture bool, listener func(Event)) func(js.Value) {
	wrapper := func(o js.Value) { listener(wrapEvent(o)) }
	n.Object.Call("addEventListener", typ, wrapper, useCapture)
	return wrapper
}

func (n *BasicNode) RemoveEventListener(typ string, useCapture bool, listener func(js.Value)) {
	n.Object.Call("removeEventListener", typ, listener, useCapture)
}

func (n *BasicNode) DispatchEvent(event Event) bool {
	return n.Object.Call("dispatchEvent", event).Bool()
}

func (n *BasicNode) BaseURI() string {
	return n.Object.Get("baseURI").String()
}

func (n *BasicNode) ChildNodes() []Node {
	return nodeListToNodes(n.Object.Get("childNodes"))
}

func (n *BasicNode) FirstChild() Node {
	return wrapNode(n.Object.Get("firstChild"))
}

func (n *BasicNode) LastChild() Node {
	return wrapNode(n.Object.Get("lastChild"))
}

func (n *BasicNode) NextSibling() Node {
	return wrapNode(n.Object.Get("nextSibling"))
}

func (n *BasicNode) NodeName() string {
	return n.Object.Get("nodeName").String()
}

func (n *BasicNode) NodeType() int {
	return n.Object.Get("nodeType").Int()
}

func (n *BasicNode) NodeValue() string {
	return toString(n.Object.Get("nodeValue"))
}

func (n *BasicNode) SetNodeValue(s string) {
	n.Object.Set("nodeValue", s)
}

func (n *BasicNode) OwnerDocument() Document {
	// FIXME implement
	panic("not implemented")
}

func (n *BasicNode) ParentNode() Node {
	return wrapNode(n.Object.Get("parentNode"))
}

func (n *BasicNode) ParentElement() Element {
	return wrapElement(n.Object.Get("parentElement"))
}

func (n *BasicNode) PreviousSibling() Node {
	return wrapNode(n.Object.Get("previousSibling"))
}

func (n *BasicNode) TextContent() string {
	return toString(n.Object.Get("textContent"))
}

func (n *BasicNode) SetTextContent(s string) {
	n.Object.Set("textContent", s)
}

func (n *BasicNode) AppendChild(newchild Node) {
	n.Object.Call("appendChild", newchild.Underlying())
}

func (n *BasicNode) CloneNode(deep bool) Node {
	return wrapNode(n.Object.Call("cloneNode", deep))
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
	return n.Object.Call("compareDocumentPosition", other.Underlying()).Int()
}

func (n *BasicNode) Contains(other Node) bool {
	return n.Object.Call("contains", other.Underlying()).Bool()
}

func (n *BasicNode) HasChildNodes() bool {
	return n.Object.Call("hasChildNodes").Bool()
}

func (n *BasicNode) InsertBefore(which Node, before Node) {
	var o interface{}
	if before != nil {
		o = before.Underlying()
	}
	n.Object.Call("insertBefore", which.Underlying(), o)
}

func (n *BasicNode) IsDefaultNamespace(s string) bool {
	return n.Object.Call("isDefaultNamespace", s).Bool()
}

func (n *BasicNode) IsEqualNode(other Node) bool {
	return n.Object.Call("isEqualNode", other.Underlying()).Bool()
}

func (n *BasicNode) LookupPrefix() string {
	return n.Object.Call("lookupPrefix").String()
}

func (n *BasicNode) LookupNamespaceURI(s string) string {
	return toString(n.Object.Call("lookupNamespaceURI", s))
}

func (n *BasicNode) Normalize() {
	n.Object.Call("normalize")
}

func (n *BasicNode) RemoveChild(other Node) {
	n.Object.Call("removeChild", other.Underlying())
}

func (n *BasicNode) ReplaceChild(newChild, oldChild Node) {
	n.Object.Call("replaceChild", newChild.Underlying(), oldChild.Underlying())
}

type Element interface {
	Node
	ParentNode
	ChildNode

	Attributes() map[string]string
	Class() *TokenList
	ID() string
	SetID(string)
	TagName() string
	GetAttribute(string) string                   // TODO can attributes only be strings?
	GetAttributeNS(ns string, name string) string // can attributes only be strings?
	GetBoundingClientRect() ClientRect
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
	InnerHTML() string
	SetInnerHTML(string)
	OuterHTML() string
	SetOuterHTML(string)
}

type ClientRect struct {
	js.Value
	Height float64 `js:"height"`
	Width  float64 `js:"width"`
	Left   float64 `js:"left"`
	Right  float64 `js:"right"`
	Top    float64 `js:"top"`
	Bottom float64 `js:"bottom"`
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
	return e.Object.Get("accessKey").String()
}

func (e *BasicHTMLElement) Dataset() map[string]string {
	o := e.Object.Get("dataset")
	data := map[string]string{}
	keys := jsKeys(o)
	for _, key := range keys {
		data[key] = o.Get(key).String()
	}
	return data
}

func (e *BasicHTMLElement) SetAccessKey(s string) {
	e.Object.Set("accessKey", s)
}

func (e *BasicHTMLElement) AccessKeyLabel() string {
	return e.Object.Get("accessKeyLabel").String()
}

func (e *BasicHTMLElement) SetAccessKeyLabel(s string) {
	e.Object.Set("accessKeyLabel", s)
}

func (e *BasicHTMLElement) ContentEditable() string {
	return e.Object.Get("contentEditable").String()
}

func (e *BasicHTMLElement) SetContentEditable(s string) {
	e.Object.Set("contentEditable", s)
}

func (e *BasicHTMLElement) IsContentEditable() bool {
	return e.Object.Get("isContentEditable").Bool()
}

func (e *BasicHTMLElement) Dir() string {
	return e.Object.Get("dir").String()
}

func (e *BasicHTMLElement) SetDir(s string) {
	e.Object.Set("dir", s)
}

func (e *BasicHTMLElement) Draggable() bool {
	return e.Object.Get("draggable").Bool()
}

func (e *BasicHTMLElement) SetDraggable(b bool) {
	e.Object.Set("draggable", b)
}

func (e *BasicHTMLElement) Lang() string {
	return e.Object.Get("lang").String()
}

func (e *BasicHTMLElement) SetLang(s string) {
	e.Object.Set("lang", s)
}

func (e *BasicHTMLElement) OffsetHeight() float64 {
	return e.Object.Get("offsetHeight").Float()
}

func (e *BasicHTMLElement) OffsetLeft() float64 {
	return e.Object.Get("offsetLeft").Float()
}

func (e *BasicHTMLElement) OffsetParent() HTMLElement {
	return wrapHTMLElement(e.Object.Get("offsetParent"))
}

func (e *BasicHTMLElement) OffsetTop() float64 {
	return e.Object.Get("offsetTop").Float()
}

func (e *BasicHTMLElement) OffsetWidth() float64 {
	return e.Object.Get("offsetWidth").Float()
}

func (e *BasicHTMLElement) Style() *CSSStyleDeclaration {
	return &CSSStyleDeclaration{e.Object.Get("style")}
}

func (e *BasicHTMLElement) TabIndex() int {
	return e.Object.Get("tabIndex").Int()
}

func (e *BasicHTMLElement) SetTabIndex(i int) {
	e.Object.Set("tabIndex", i)
}

func (e *BasicHTMLElement) Title() string {
	return e.Object.Get("title").String()
}

func (e *BasicHTMLElement) SetTitle(s string) {
	e.Object.Set("title", s)
}

func (e *BasicHTMLElement) Blur() {
	e.Object.Call("blur")
}

func (e *BasicHTMLElement) Click() {
	e.Object.Call("click")
}

func (e *BasicHTMLElement) Focus() {
	e.Object.Call("focus")
}

// Type BasicElement implements the Element interface and is embedded
// by concrete element types and HTML element types.
type BasicElement struct {
	*BasicNode
}

func (e *BasicElement) Attributes() map[string]string {
	o := e.Object.Get("attributes")
	attrs := map[string]string{}
	length := o.Get("length").Int()
	for i := 0; i < length; i++ {
		item := o.Call("item", i)
		attrs[item.Get("name").String()] = item.Get("value").String()
	}
	return attrs
}

func (e *BasicElement) GetBoundingClientRect() ClientRect {
	obj := e.Object.Call("getBoundingClientRect")
	return ClientRect{Value: obj}
}

func (e *BasicElement) PreviousElementSibling() Element {
	return wrapElement(e.Object.Get("previousElementSibling"))
}

func (e *BasicElement) NextElementSibling() Element {
	return wrapElement(e.Object.Get("nextElementSibling"))
}

func (e *BasicElement) Class() *TokenList {
	return &TokenList{dtl: e.Object.Get("classList"), o: e.Object, sa: "className"}
}

// SetClass sets the element's className attribute to s. Consider
// using the Class method instead.
func (e *BasicElement) SetClass(s string) {
	e.Object.Set("className", s)
}

func (e *BasicElement) ID() string {
	return e.Object.Get("id").String()
}

func (e *BasicElement) SetID(s string) {
	e.Object.Set("id", s)
}

func (e *BasicElement) TagName() string {
	return e.Object.Get("tagName").String()
}

func (e *BasicElement) GetAttribute(name string) string {
	return toString(e.Object.Call("getAttribute", name))
}

func (e *BasicElement) GetAttributeNS(ns string, name string) string {
	return toString(e.Object.Call("getAttributeNS", ns, name))
}

func (e *BasicElement) GetElementsByClassName(s string) []Element {
	return nodeListToElements(e.Object.Call("getElementsByClassName", s))
}

func (e *BasicElement) GetElementsByTagName(s string) []Element {
	return nodeListToElements(e.Object.Call("getElementsByTagName", s))
}

func (e *BasicElement) GetElementsByTagNameNS(ns string, name string) []Element {
	return nodeListToElements(e.Object.Call("getElementsByTagNameNS", ns, name))
}

func (e *BasicElement) HasAttribute(s string) bool {
	return e.Object.Call("hasAttribute", s).Bool()
}

func (e *BasicElement) HasAttributeNS(ns string, name string) bool {
	return e.Object.Call("hasAttributeNS", ns, name).Bool()
}

func (e *BasicElement) QuerySelector(s string) Element {
	return wrapElement(e.Object.Call("querySelector", s))
}

func (e *BasicElement) QuerySelectorAll(s string) []Element {
	return nodeListToElements(e.Object.Call("querySelectorAll", s))
}

func (e *BasicElement) RemoveAttribute(s string) {
	e.Object.Call("removeAttribute", s)
}

func (e *BasicElement) RemoveAttributeNS(ns string, name string) {
	e.Object.Call("removeAttributeNS", ns, name)
}

func (e *BasicElement) SetAttribute(name string, value string) {
	e.Object.Call("setAttribute", name, value)
}

func (e *BasicElement) SetAttributeNS(ns string, name string, value string) {
	e.Object.Call("setAttributeNS", ns, name, value)
}

func (e *BasicElement) InnerHTML() string {
	return e.Object.Get("innerHTML").String()
}

func (e *BasicElement) SetInnerHTML(s string) {
	e.Object.Set("innerHTML", s)
}

func (e *BasicElement) OuterHTML() string {
	return e.Object.Get("outerHTML").String()
}

func (e *BasicElement) SetOuterHTML(s string) {
	e.Object.Set("outerHTML", s)
}

type HTMLAnchorElement struct {
	*BasicHTMLElement
	*URLUtils
	HrefLang string `js:"hreflang"`
	Media    string `js:"media"`
	TabIndex int    `js:"tabIndex"`
	Target   string `js:"target"`
	Text     string `js:"text"`
	Type     string `js:"type"`
}

func (e *HTMLAnchorElement) Rel() *TokenList {
	return &TokenList{dtl: e.Object.Get("relList"), o: e.Object, sa: "rel"}
}

type HTMLAppletElement struct {
	*BasicHTMLElement
	Alt      string `js:"alt"`
	Coords   string `js:"coords"`
	HrefLang string `js:"hreflang"`
	Media    string `js:"media"`
	Search   string `js:"search"`
	Shape    string `js:"shape"`
	TabIndex int    `js:"tabIndex"`
	Target   string `js:"target"`
	Type     string `js:"type"`
}

func (e *HTMLAppletElement) Rel() *TokenList {
	return &TokenList{dtl: e.Object.Get("relList"), o: e.Object, sa: "rel"}
}

type HTMLAreaElement struct {
	*BasicHTMLElement
	*URLUtils
	Alt      string `js:"alt"`
	Coords   string `js:"coords"`
	HrefLang string `js:"hreflang"`
	Media    string `js:"media"`
	Search   string `js:"search"`
	Shape    string `js:"shape"`
	TabIndex int    `js:"tabIndex"`
	Target   string `js:"target"`
	Type     string `js:"type"`
}

func (e *HTMLAreaElement) Rel() *TokenList {
	return &TokenList{dtl: e.Object.Get("relList"), o: e.Object, sa: "rel"}
}

type HTMLAudioElement struct{ *HTMLMediaElement }

type HTMLBRElement struct{ *BasicHTMLElement }

type HTMLBaseElement struct{ *BasicHTMLElement }

func (e *HTMLBaseElement) Href() string {
	return e.Object.Get("href").String()
}

func (e *HTMLBaseElement) Target() string {
	return e.Object.Get("target").String()
}

type HTMLBodyElement struct{ *BasicHTMLElement }

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
	return getForm(e.Object)
}

func (e *HTMLButtonElement) Labels() []*HTMLLabelElement {
	return getLabels(e.Object)
}

func (e *HTMLButtonElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLButtonElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLButtonElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

type HTMLCanvasElement struct {
	*BasicHTMLElement
	Height int `js:"height"`
	Width  int `js:"width"`
}

type CanvasRenderingContext2D struct {
	js.Value

	// Colors, Styles, and Shadows

	FillStyle     string `js:"fillStyle"`
	StrokeStyle   string `js:"strokeStyle"`
	ShadowColor   string `js:"shadowColor"`
	ShadowBlur    int    `js:"shadowBlur"`
	ShadowOffsetX int    `js:"shadowOffsetX"`
	ShadowOffsetY int    `js:"shadowOffsetY"`

	// Line Styles

	LineCap    string `js:"lineCap"`
	LineJoin   string `js:"lineJoin"`
	LineWidth  int    `js:"lineWidth"`
	MiterLimit int    `js:"miterLimit"`

	// Text

	Font         string `js:"font"`
	TextAlign    string `js:"textAlign"`
	TextBaseline string `js:"textBaseline"`

	// Compositing

	GlobalAlpha              float64 `js:"globalAlpha"`
	GlobalCompositeOperation string  `js:"globalCompositeOperation"`
}

type ImageData struct {
	js.Value

	Width  int      `js:"width"`
	Height int      `js:"height"`
	Data   js.Value `js:"data"`
}

func (m *ImageData) ColorModel() color.Model { return color.NRGBAModel }

func (m *ImageData) Bounds() image.Rectangle {
	return image.Rect(0, 0, m.Width, m.Height)
}

func (m *ImageData) At(x, y int) color.Color {
	return m.NRGBAAt(x, y)
}

func (m *ImageData) NRGBAAt(x, y int) color.NRGBA {
	if x < 0 || x >= m.Width ||
		y < 0 || y >= m.Height {
		return color.NRGBA{}
	}
	i := (y*m.Width + x) * 4
	return color.NRGBA{
		R: uint8(m.Data.Index(i + 0).Int()),
		G: uint8(m.Data.Index(i + 1).Int()),
		B: uint8(m.Data.Index(i + 2).Int()),
		A: uint8(m.Data.Index(i + 3).Int()),
	}
}

func (m *ImageData) Set(x, y int, c color.Color) {
	if x < 0 || x >= m.Width ||
		y < 0 || y >= m.Height {
		return
	}
	c1 := color.NRGBAModel.Convert(c).(color.NRGBA)
	i := (y*m.Width + x) * 4
	m.Data.SetIndex(i+0, c1.R)
	m.Data.SetIndex(i+1, c1.G)
	m.Data.SetIndex(i+2, c1.B)
	m.Data.SetIndex(i+3, c1.A)
}

func (m *ImageData) SetNRGBA(x, y int, c color.NRGBA) {
	if x < 0 || x >= m.Width ||
		y < 0 || y >= m.Height {
		return
	}
	i := (y*m.Width + x) * 4
	m.Data.SetIndex(i+0, c.R)
	m.Data.SetIndex(i+1, c.G)
	m.Data.SetIndex(i+2, c.B)
	m.Data.SetIndex(i+3, c.A)
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

	Width                    float64 `js:"width"`
	ActualBoundingBoxLeft    float64 `js:"actualBoundingBoxLeft"`
	ActualBoundingBoxRight   float64 `js:"actualBoundingBoxRight"`
	FontBoundingBoxAscent    float64 `js:"fontBoundingBoxAscent"`
	FontBoundingBoxDescent   float64 `js:"fontBoundingBoxDescent"`
	ActualBoundingBoxAscent  float64 `js:"actualBoundingBoxAscent"`
	ActualBoundingBoxDescent float64 `js:"actualBoundingBoxDescent"`
	EmHeightAscent           float64 `js:"emHeightAscent"`
	EmHeightDescent          float64 `js:"emHeightDescent"`
	HangingBaseline          float64 `js:"hangingBaseline"`
	AlphabeticBaseline       float64 `js:"alphabeticBaseline"`
	IdeographicBaseline      float64 `js:"ideographicBaseline"`
}

// Creating canvas 2d context

func (e *HTMLCanvasElement) GetContext2d() *CanvasRenderingContext2D {
	ctx := e.GetContext("2d")
	return &CanvasRenderingContext2D{Value: ctx}
}

func (e *HTMLCanvasElement) GetContext(param string) js.Value {
	return e.Object.Call("getContext", param)
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
	Value string `js:"value"`
}

type HTMLDataListElement struct{ *BasicHTMLElement }

func (e *HTMLDataListElement) Options() []*HTMLOptionElement {
	return getOptions(e.Object, "options")
}

type HTMLDirectoryElement struct{ *BasicHTMLElement }
type HTMLDivElement struct{ *BasicHTMLElement }

type HTMLEmbedElement struct {
	*BasicHTMLElement
	Src   string `js:"src"`
	Type  string `js:"type"`
	Width string `js:"width"`
}

type HTMLFieldSetElement struct {
	*BasicHTMLElement
	Disabled          bool   `js:"disabled"`
	Name              string `js:"name"`
	Type              string `js:"type"`
	ValidationMessage string `js:"validationMessage"`
	WillValidate      bool   `js:"willValidate"`
}

func (e *HTMLFieldSetElement) Elements() []HTMLElement {
	return nodeListToHTMLElements(e.Object.Get("elements"))
}

func (e *HTMLFieldSetElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLFieldSetElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLFieldSetElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLFieldSetElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

type HTMLFontElement struct{ *BasicHTMLElement }

type HTMLFormElement struct {
	*BasicHTMLElement
	AcceptCharset string `js:"acceptCharset"`
	Action        string `js:"action"`
	Autocomplete  string `js:"autocomplete"`
	Encoding      string `js:"encoding"`
	Enctype       string `js:"enctype"`
	Length        int    `js:"length"`
	Method        string `js:"method"`
	Name          string `js:"name"`
	NoValidate    bool   `js:"noValidate"`
	Target        string `js:"target"`
}

func (e *HTMLFormElement) Elements() []HTMLElement {
	return nodeListToHTMLElements(e.Object.Get("elements"))
}

func (e *HTMLFormElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLFormElement) Submit() {
	e.Object.Call("submit")
}

func (e *HTMLFormElement) Reset() {
	e.Object.Call("reset")
}

func (e *HTMLFormElement) Item(index int) HTMLElement {
	return wrapHTMLElement(e.Object.Call("item", index))
}

func (e *HTMLFormElement) NamedItem(name string) HTMLElement {
	return wrapHTMLElement(e.Object.Call("namedItem", name))
}

type HTMLFrameElement struct{ *BasicHTMLElement }
type HTMLFrameSetElement struct{ *BasicHTMLElement }
type HTMLHRElement struct{ *BasicHTMLElement }
type HTMLHeadElement struct{ *BasicHTMLElement }
type HTMLHeadingElement struct{ *BasicHTMLElement }
type HTMLHtmlElement struct{ *BasicHTMLElement }

type HTMLIFrameElement struct {
	*BasicHTMLElement
	Width    string `js:"width"`
	Height   string `js:"height"`
	Name     string `js:"name"`
	Src      string `js:"src"`
	SrcDoc   string `js:"srcdoc"`
	Seamless bool   `js:"seamless"`
	// TODO sandbox attribute
}

func (e *HTMLIFrameElement) ContentDocument() Document {
	return wrapDocument(e.Object.Get("contentDocument"))
}

func (e *HTMLIFrameElement) ContentWindow() Window {
	return &window{e.Object.Get("contentWindow")}
}

type HTMLImageElement struct {
	*BasicHTMLElement
	Complete      bool   `js:"complete"`
	CrossOrigin   string `js:"crossOrigin"`
	Height        int    `js:"height"`
	IsMap         bool   `js:"isMap"`
	NaturalHeight int    `js:"naturalHeight"`
	NaturalWidth  int    `js:"naturalWidth"`
	Src           string `js:"src"`
	UseMap        string `js:"useMap"`
	Width         int    `js:"width"`
	// TODO constructor
}

type HTMLInputElement struct {
	*BasicHTMLElement
	Accept             string    `js:"accept"`
	Alt                string    `js:"alt"`
	Autocomplete       string    `js:"autocomplete"`
	Autofocus          bool      `js:"autofocus"`
	Checked            bool      `js:"checked"`
	DefaultChecked     bool      `js:"defaultChecked"`
	DefaultValue       string    `js:"defaultValue"`
	DirName            string    `js:"dirName"`
	Disabled           bool      `js:"disabled"`
	FormAction         string    `js:"formAction"`
	FormEncType        string    `js:"formEncType"`
	FormMethod         string    `js:"formMethod"`
	FormNoValidate     bool      `js:"formNoValidate"`
	FormTarget         string    `js:"formTarget"`
	Height             string    `js:"height"`
	Indeterminate      bool      `js:"indeterminate"`
	Max                string    `js:"max"`
	MaxLength          int       `js:"maxLength"`
	Min                string    `js:"min"`
	Multiple           bool      `js:"multiple"`
	Name               string    `js:"name"`
	Pattern            string    `js:"pattern"`
	Placeholder        string    `js:"placeholder"`
	ReadOnly           bool      `js:"readOnly"`
	Required           bool      `js:"required"`
	SelectionDirection string    `js:"selectionDirection"`
	SelectionEnd       int       `js:"selectionEnd"`
	SelectionStart     int       `js:"selectionStart"`
	Size               int       `js:"size"`
	Src                string    `js:"src"`
	Step               string    `js:"step"`
	TabIndex           int       `js:"tabIndex"`
	Type               string    `js:"type"`
	ValidationMessage  string    `js:"validationMessage"`
	Value              string    `js:"value"`
	ValueAsDate        time.Time `js:"valueAsDate"`
	ValueAsNumber      float64   `js:"valueAsNumber"`
	Width              string    `js:"width"`
	WillValidate       bool      `js:"willValidate"`
}

// File represents files as can be obtained from file choosers or drag
// and drop. The dom package does not define any methods on File nor
// does it provide access to the blob or a way to read it.
type File struct {
	js.Value
}

func (e *HTMLInputElement) Files() []*File {
	files := e.Object.Get("files")
	out := make([]*File, files.Get("length").Int())
	for i := range out {
		out[i] = &File{files.Call("item", i)}
	}
	return out
}

func (e *HTMLInputElement) List() *HTMLDataListElement {
	list := wrapHTMLElement(e.Object.Get("list"))
	if list == nil {
		return nil
	}
	return list.(*HTMLDataListElement)
}

func (e *HTMLInputElement) Labels() []*HTMLLabelElement {
	return getLabels(e.Object)
}

func (e *HTMLInputElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLInputElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLInputElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLInputElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

func (e *HTMLInputElement) Select() {
	e.Object.Call("select")
}

func (e *HTMLInputElement) SetSelectionRange(start, end int, direction string) {
	e.Object.Call("setSelectionRange", start, end, direction)
}

func (e *HTMLInputElement) StepDown(n int) error {
	return callRecover(e.Object, "stepDown", n)
}

func (e *HTMLInputElement) StepUp(n int) error {
	return callRecover(e.Object, "stepUp", n)
}

type HTMLKeygenElement struct {
	*BasicHTMLElement
	Autofocus         bool   `js:"autofocus"`
	Challenge         string `js:"challenge"`
	Disabled          bool   `js:"disabled"`
	Keytype           string `js:"keytype"`
	Name              string `js:"name"`
	Type              string `js:"type"`
	ValidationMessage string `js:"validationMessage"`
	WillValidate      bool   `js:"willValidate"`
}

func (e *HTMLKeygenElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLKeygenElement) Labels() []*HTMLLabelElement {
	return getLabels(e.Object)
}

func (e *HTMLKeygenElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLKeygenElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLKeygenElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

type HTMLLIElement struct {
	*BasicHTMLElement
	Value int `js:"value"`
}

type HTMLLabelElement struct {
	*BasicHTMLElement
	For string `js:"htmlFor"`
}

func (e *HTMLLabelElement) Control() HTMLElement {
	return wrapHTMLElement(e.Object.Get("control"))
}

func (e *HTMLLabelElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

type HTMLLegendElement struct{ *BasicHTMLElement }

func (e *HTMLLegendElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

type HTMLLinkElement struct {
	*BasicHTMLElement
	Disabled bool   `js:"disabled"`
	Href     string `js:"href"`
	HrefLang string `js:"hrefLang"`
	Media    string `js:"media"`
	Type     string `js:"type"`
}

func (e *HTMLLinkElement) Rel() *TokenList {
	return &TokenList{dtl: e.Object.Get("relList"), o: e.Object, sa: "rel"}
}

func (e *HTMLLinkElement) Sizes() *TokenList {
	return &TokenList{dtl: e.Object.Get("sizes"), o: e.Object}
}

func (e *HTMLLinkElement) Sheet() StyleSheet {
	// FIXME implement
	panic("not implemented")
}

type HTMLMapElement struct {
	*BasicHTMLElement
	Name string `js:"name"`
}

func (e *HTMLMapElement) Areas() []*HTMLAreaElement {
	areas := nodeListToElements(e.Object.Get("areas"))
	out := make([]*HTMLAreaElement, len(areas))
	for i, area := range areas {
		out[i] = area.(*HTMLAreaElement)
	}
	return out
}

func (e *HTMLMapElement) Images() []HTMLElement {
	return nodeListToHTMLElements(e.Object.Get("areas"))
}

type HTMLMediaElement struct {
	*BasicHTMLElement
	Paused bool `js:"paused"`
}

func (e *HTMLMediaElement) Play() {
	e.Object.Call("play")
}

func (e *HTMLMediaElement) Pause() {
	e.Object.Call("pause")
}

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
	return getLabels(e.Object)
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

type HTMLObjectElement struct {
	*BasicHTMLElement
	Data              string `js:"data"`
	Height            string `js:"height"`
	Name              string `js:"name"`
	TabIndex          int    `js:"tabIndex"`
	Type              string `js:"type"`
	TypeMustMatch     bool   `js:"typeMustMatch"`
	UseMap            string `js:"useMap"`
	ValidationMessage string `js:"validationMessage"`
	With              string `js:"with"`
	WillValidate      bool   `js:"willValidate"`
}

func (e *HTMLObjectElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLObjectElement) ContentDocument() Document {
	return wrapDocument(e.Object.Get("contentDocument"))
}

func (e *HTMLObjectElement) ContentWindow() Window {
	return &window{e.Object.Get("contentWindow")}
}

func (e *HTMLObjectElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLObjectElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLObjectElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

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
	return getForm(e.Object)
}

type HTMLOutputElement struct {
	*BasicHTMLElement
	DefaultValue      string `js:"defaultValue"`
	Name              string `js:"name"`
	Type              string `js:"type"`
	ValidationMessage string `js:"validationMessage"`
	Value             string `js:"value"`
	WillValidate      bool   `js:"willValidate"`
}

func (e *HTMLOutputElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLOutputElement) Labels() []*HTMLLabelElement {
	return getLabels(e.Object)
}

func (e *HTMLOutputElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLOutputElement) For() *TokenList {
	return &TokenList{dtl: e.Object.Get("htmlFor"), o: e.Object}
}

func (e *HTMLOutputElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLOutputElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

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
	return getLabels(e.Object)
}

type HTMLQuoteElement struct {
	*BasicHTMLElement
	Cite string `js:"cite"`
}

type HTMLScriptElement struct {
	*BasicHTMLElement
	Type    string `js:"type"`
	Src     string `js:"src"`
	Charset string `js:"charset"`
	Async   bool   `js:"async"`
	Defer   bool   `js:"defer"`
	Text    string `js:"text"`
}

type HTMLSelectElement struct {
	*BasicHTMLElement
	Autofocus         bool   `js:"autofocus"`
	Disabled          bool   `js:"disabled"`
	Length            int    `js:"length"`
	Multiple          bool   `js:"multiple"`
	Name              string `js:"name"`
	Required          bool   `js:"required"`
	SelectedIndex     int    `js:"selectedIndex"`
	Size              int    `js:"size"`
	Type              string `js:"type"`
	ValidationMessage string `js:"validationMessage"`
	Value             string `js:"value"`
	WillValidate      bool   `js:"willValidate"`
}

func (e *HTMLSelectElement) Labels() []*HTMLLabelElement {
	return getLabels(e.Object)
}

func (e *HTMLSelectElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLSelectElement) Options() []*HTMLOptionElement {
	return getOptions(e.Object, "options")
}

func (e *HTMLSelectElement) SelectedOptions() []*HTMLOptionElement {
	return getOptions(e.Object, "selectedOptions")
}

func (e *HTMLSelectElement) Item(index int) *HTMLOptionElement {
	el := wrapHTMLElement(e.Object.Call("item", index))
	if el == nil {
		return nil
	}
	return el.(*HTMLOptionElement)
}

func (e *HTMLSelectElement) NamedItem(name string) *HTMLOptionElement {
	el := wrapHTMLElement(e.Object.Call("namedItem", name))
	if el == nil {
		return nil
	}
	return el.(*HTMLOptionElement)
}

// TODO(dominikh): Not implementing Add or Remove for now. For one,
// Add with "before" behaves weird when dealing with optgroups. Also,
// there's already InsertBefore and RemoveChild which can be used
// instead.

func (e *HTMLSelectElement) Validity() *ValidityState {
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLSelectElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLSelectElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

type HTMLSourceElement struct {
	*BasicHTMLElement
	Media string `js:"media"`
	Src   string `js:"src"`
	Type  string `js:"type"`
}

type HTMLSpanElement struct{ *BasicHTMLElement }
type HTMLStyleElement struct{ *BasicHTMLElement }
type HTMLTableCaptionElement struct{ *BasicHTMLElement }

type HTMLTableCellElement struct {
	*BasicHTMLElement
	ColSpan   int `js:"colSpan"`
	RowSpan   int `js:"rowSpan"`
	CellIndex int `js:"cellIndex"`
	// TODO headers
}

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

type HTMLTableRowElement struct {
	*BasicHTMLElement
	RowIndex        int `js:"rowIndex"`
	SectionRowIndex int `js:"sectionRowIndex"`
}

func (e *HTMLTableRowElement) Cells() []*HTMLTableCellElement {
	cells := nodeListToElements(e.Object.Get("cells"))
	out := make([]*HTMLTableCellElement, len(cells))
	for i, cell := range cells {
		out[i] = cell.(*HTMLTableCellElement)
	}
	return out
}

func (e *HTMLTableRowElement) InsertCell(index int) *HTMLTableCellElement {
	return wrapHTMLElement(e.Object.Call("insertCell", index)).(*HTMLTableCellElement)
}

func (e *HTMLTableRowElement) DeleteCell(index int) {
	// FIXME exception handling/check that index is in bounds
	e.Object.Call("deleteCell", index)
}

type HTMLTableSectionElement struct{ *BasicHTMLElement }

func (e *HTMLTableSectionElement) Rows() []*HTMLTableRowElement {
	rows := nodeListToElements(e.Object.Get("rows"))
	out := make([]*HTMLTableRowElement, len(rows))
	for i, row := range rows {
		out[i] = row.(*HTMLTableRowElement)
	}
	return out
}

func (e *HTMLTableSectionElement) DeleteRow(index int) {
	// FIXME exception handling/check that index is in bounds
	e.Object.Call("deleteRow", index)
}

func (e *HTMLTableSectionElement) InsertRow(index int) *HTMLTableRowElement {
	return wrapHTMLElement(e.Object.Call("insertRow", index)).(*HTMLTableRowElement)
}

type HTMLTextAreaElement struct {
	*BasicHTMLElement
	Autocomplete       string `js:"autocomplete"`
	Autofocus          bool   `js:"autofocus"`
	Cols               int    `js:"cols"`
	DefaultValue       string `js:"defaultValue"`
	DirName            string `js:"dirName"`
	Disabled           bool   `js:"disabled"`
	MaxLength          int    `js:"maxLength"`
	Name               string `js:"name"`
	Placeholder        string `js:"placeholder"`
	ReadOnly           bool   `js:"readOnly"`
	Required           bool   `js:"required"`
	Rows               int    `js:"rows"`
	SelectionDirection string `js:"selectionDirection"`
	SelectionStart     int    `js:"selectionStart"`
	SelectionEnd       int    `js:"selectionEnd"`
	TabIndex           int    `js:"tabIndex"`
	TextLength         int    `js:"textLength"`
	Type               string `js:"type"`
	ValidationMessage  string `js:"validationMessage"`
	Value              string `js:"value"`
	WillValidate       bool   `js:"willValidate"`
	Wrap               string `js:"wrap"`
}

func (e *HTMLTextAreaElement) Form() *HTMLFormElement {
	return getForm(e.Object)
}

func (e *HTMLTextAreaElement) Labels() []*HTMLLabelElement {
	return getLabels(e.Object)
}

func (e *HTMLTextAreaElement) Validity() *ValidityState {
	// TODO replace with a field once GopherJS supports that
	return &ValidityState{Value: e.Object.Get("validity")}
}

func (e *HTMLTextAreaElement) CheckValidity() bool {
	return e.Object.Call("checkValidity").Bool()
}

func (e *HTMLTextAreaElement) SetCustomValidity(s string) {
	e.Object.Call("setCustomValidity", s)
}

func (e *HTMLTextAreaElement) Select() {
	e.Object.Call("select")
}

func (e *HTMLTextAreaElement) SetSelectionRange(start, end int, direction string) {
	e.Object.Call("setSelectionRange", start, end, direction)
}

type HTMLTimeElement struct {
	*BasicHTMLElement
	DateTime string `js:"dateTime"`
}

type HTMLTitleElement struct {
	*BasicHTMLElement
	Text string `js:"text"`
}

// TextTrack represents text track data for <track> elements. It does
// not currently provide any methods or attributes and it hasn't been
// decided yet whether they will be added to this package or a
// separate package.
type TextTrack struct{ js.Value }

type HTMLTrackElement struct {
	*BasicHTMLElement
	Kind       string `js:"kind"`
	Src        string `js:"src"`
	Srclang    string `js:"srclang"`
	Label      string `js:"label"`
	Default    bool   `js:"default"`
	ReadyState int    `js:"readyState"`
}

func (e *HTMLTrackElement) Track() *TextTrack {
	return &TextTrack{e.Object.Get("track")}
}

type HTMLUListElement struct{ *BasicHTMLElement }
type HTMLUnknownElement struct{ *BasicHTMLElement }

type HTMLVideoElement struct{ *HTMLMediaElement }

type ValidityState struct {
	js.Value
	CustomError     bool `js:"customError"`
	PatternMismatch bool `js:"patternMismatch"`
	RangeOverflow   bool `js:"rangeOverflow"`
	RangeUnderflow  bool `js:"rangeUnderflow"`
	StepMismatch    bool `js:"stepMismatch"`
	TooLong         bool `js:"tooLong"`
	TypeMismatch    bool `js:"typeMismatch"`
	Valid           bool `js:"valid"`
	ValueMissing    bool `js:"valueMissing"`
}

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

func (css *CSSStyleDeclaration) RemoveProperty(name string) {
	css.Call("removeProperty", name)
}

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
