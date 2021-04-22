/*
 * Copyright 2020 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcfynego

import (
	"aletheiaware.com/aliasgo"
	"aletheiaware.com/bcclientgo"
	"aletheiaware.com/bcfynego/storage"
	"aletheiaware.com/bcfynego/ui"
	accountui "aletheiaware.com/bcfynego/ui/account"
	"aletheiaware.com/bcfynego/ui/data"
	"aletheiaware.com/bcgo"
	"aletheiaware.com/bcgo/account"
	"aletheiaware.com/bcgo/node"
	"aletheiaware.com/cryptogo"
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

type BCFyne interface {
	App() fyne.App
	Window() fyne.Window
	AddOnKeysExported(func(string))
	AddOnKeysImported(func(string))
	AddOnSignedIn(func(bcgo.Account))
	AddOnSignedUp(func(bcgo.Account))
	AddOnSignedOut(func())
	DeleteKeys(bcclientgo.BCClient, bcgo.Account)
	ExportKeys(bcclientgo.BCClient, bcgo.Account)
	Logo() fyne.CanvasObject
	Account(bcclientgo.BCClient) (bcgo.Account, error)
	Node(bcclientgo.BCClient) (bcgo.Node, error)
	ShowAccessDialog(bcclientgo.BCClient, func(bcgo.Account))
	ShowAccount(bcclientgo.BCClient)
	ShowError(error)
	ShowURI(bcclientgo.BCClient, fyne.URI)
	SignOut(bcclientgo.BCClient)
}

type bcFyne struct {
	app            fyne.App
	window         fyne.Window
	onKeysExported []func(string)
	onKeysImported []func(string)
	onSignedIn     []func(bcgo.Account)
	onSignedUp     []func(bcgo.Account)
	onSignedOut    []func()
}

func NewBCFyne(a fyne.App, w fyne.Window) BCFyne {
	return &bcFyne{
		app:    a,
		window: w,
	}
}

func (f *bcFyne) App() fyne.App {
	return f.app
}

func (f *bcFyne) Window() fyne.Window {
	return f.window
}

func (f *bcFyne) AddOnKeysExported(callback func(string)) {
	f.onKeysExported = append(f.onKeysExported, callback)
}

func (f *bcFyne) AddOnKeysImported(callback func(string)) {
	f.onKeysImported = append(f.onKeysImported, callback)
}

func (f *bcFyne) AddOnSignedIn(callback func(bcgo.Account)) {
	f.onSignedIn = append(f.onSignedIn, callback)
}

func (f *bcFyne) AddOnSignedUp(callback func(bcgo.Account)) {
	f.onSignedUp = append(f.onSignedUp, callback)
}

func (f *bcFyne) AddOnSignedOut(callback func()) {
	f.onSignedOut = append(f.onSignedOut, callback)
}

func (f *bcFyne) ExistingAccount(client bcclientgo.BCClient, alias string, password []byte, callback func(bcgo.Account)) {
	rootDir, err := client.Root()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Get key store
	keystore, err := bcgo.KeyDirectory(rootDir)
	if err != nil {
		f.ShowError(err)
		return
	}
	// Get private key
	key, err := cryptogo.RSAPrivateKey(keystore, alias, password)
	if err != nil {
		f.ShowError(err)
		return
	}
	account := account.NewRSA(alias, key)
	if c := callback; c != nil {
		c(account)
	}
}

func (f *bcFyne) Account(client bcclientgo.BCClient) (bcgo.Account, error) {
	if !client.HasAccount() {
		var wg sync.WaitGroup
		wg.Add(1)
		go f.ShowAccessDialog(client, func(a bcgo.Account) {
			client.SetAccount(a)
			wg.Done()
		})
		wg.Wait()
	}
	return client.Account()
}

func (f *bcFyne) Node(client bcclientgo.BCClient) (bcgo.Node, error) {
	if !client.HasNode() {
		account, err := f.Account(client)
		if err != nil {
			return nil, err
		}
		cache, err := client.Cache()
		if err != nil {
			return nil, err
		}
		network, err := client.Network()
		if err != nil {
			return nil, err
		}
		client.SetNode(node.New(account, cache, network))
	}
	return client.Node()
}

func (f *bcFyne) Logo() fyne.CanvasObject {
	return &canvas.Image{
		Resource: data.Logo,
		//FillMode: canvas.ImageFillContain,
		FillMode: canvas.ImageFillOriginal,
	}
}

func (f *bcFyne) NewAccount(client bcclientgo.BCClient, alias string, password []byte, callback func(bcgo.Account)) {
	// Show Progress Dialog
	progress := dialog.NewProgressInfinite("Creating", "Creating "+alias, f.window)
	progress.Show()
	defer progress.Hide()

	rootDir, err := client.Root()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Get key store
	keystore, err := bcgo.KeyDirectory(rootDir)
	if err != nil {
		f.ShowError(err)
		return
	}
	// Create private key
	key, err := cryptogo.CreateRSAPrivateKey(keystore, alias, password)
	if err != nil {
		f.ShowError(err)
		return
	}
	account := account.NewRSA(alias, key)
	cache, err := client.Cache()
	if err != nil {
		f.ShowError(err)
		return
	}
	network, err := client.Network()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Create node
	node := node.New(account, cache, network)

	{
		// Show Progress Dialog
		progress := dialog.NewProgress("Registering", "Registering "+alias, f.window)
		progress.Show()
		listener := &ui.ProgressMiningListener{Func: progress.SetValue}

		// Register Alias
		err := aliasgo.Register(node, listener)

		// Hide Progress Dialog
		progress.Hide()

		if err != nil {
			f.ShowError(err)
			return
		}
	}

	if c := callback; c != nil {
		c(account)
	}
}

func (f *bcFyne) ShowAccessDialog(client bcclientgo.BCClient, callback func(bcgo.Account)) {
	signIn := accountui.NewSignIn()
	importKey := accountui.NewImportKey()
	signUp := accountui.NewSignUp()
	accordion := widget.NewAccordion(
		&widget.AccordionItem{Title: "Sign In", Detail: signIn.CanvasObject(), Open: true},
		widget.NewAccordionItem("Import Keys", importKey.CanvasObject()),
		widget.NewAccordionItem("Sign Up", signUp.CanvasObject()),
	)
	tos := &widget.Hyperlink{Text: "Terms of Service"}
	tos.SetURLFromString("https://aletheiaware.com/terms-of-service.html")
	pp := &widget.Hyperlink{Text: "Privacy Policy", Alignment: fyne.TextAlignTrailing}
	pp.SetURLFromString("https://aletheiaware.com/privacy-policy.html")
	contents := container.NewVBox()
	if !bcgo.IsLive() {
		contents.Add(ui.NewTestModeSign())
	}
	contents.Add(accordion)
	contents.Add(container.NewMax(
		&canvas.Image{
			Resource: data.AW,
			FillMode: canvas.ImageFillContain,
		},
		container.NewGridWithColumns(2, tos, pp),
	))
	d := dialog.NewCustom("Account Access", "Cancel", contents, f.window)

	signInAction := func() {
		d.Hide()

		alias := signIn.Alias.Text
		password := []byte(signIn.Password.Text)
		if len(password) < cryptogo.MIN_PASSWORD {
			f.ShowError(cryptogo.ErrPasswordTooShort{Size: len(password), Min: cryptogo.MIN_PASSWORD})
			return
		}
		f.ExistingAccount(client, alias, password, func(account bcgo.Account) {
			if c := callback; c != nil {
				c(account)
			}
			for _, c := range f.onSignedIn {
				c(account)
			}
		})
	}
	signIn.Alias.OnSubmitted = func(string) {
		f.window.Canvas().Focus(signIn.Password)
	}
	signIn.Password.OnSubmitted = func(string) {
		signInAction()
	}
	signIn.SignInButton.OnTapped = signInAction
	importKeyAction := func() {
		d.Hide()

		host := bcgo.BCWebsite()
		alias := importKey.Alias.Text
		access := importKey.Access.Text

		// Show Progress Dialog
		progress := dialog.NewProgress("Importing Keys", fmt.Sprintf("Importing %s from %s", alias, host), f.window)
		progress.Show()

		err := client.ImportKeys(host, alias, access)

		// Hide Progress Dialog
		progress.Hide()

		if err != nil {
			f.ShowError(err)
			return
		}

		for _, c := range f.onKeysImported {
			c(alias)
		}

		authentication := accountui.NewAuthentication(alias)

		contents := container.NewVBox()
		if !bcgo.IsLive() {
			contents.Add(ui.NewTestModeSign())
		}
		contents.Add(widget.NewLabel(fmt.Sprintf("Keys for %s successfully imported from %s.\nAuthenticate to continue", alias, host)))
		contents.Add(authentication.CanvasObject())
		d := dialog.NewCustom("Keys Imported", "Cancel", contents, f.window)

		authenticateAction := func() {
			d.Hide()

			password := []byte(authentication.Password.Text)
			if len(password) < cryptogo.MIN_PASSWORD {
				f.ShowError(cryptogo.ErrPasswordTooShort{Size: len(password), Min: cryptogo.MIN_PASSWORD})
				return
			}
			f.ExistingAccount(client, alias, password, func(account bcgo.Account) {
				if c := callback; c != nil {
					c(account)
				}
				for _, c := range f.onSignedIn {
					c(account)
				}
			})
		}
		authentication.Password.OnSubmitted = func(string) {
			authenticateAction()
		}
		authentication.AuthenticateButton.OnTapped = authenticateAction

		// Show Success Dialog
		d.Show()
		d.Resize(ui.DialogSize)
	}
	importKey.Alias.OnSubmitted = func(string) {
		f.window.Canvas().Focus(importKey.Access)
	}
	importKey.Access.OnSubmitted = func(string) {
		importKeyAction()
	}
	importKey.ImportKeyButton.OnTapped = importKeyAction
	signUpAction := func() {
		d.Hide()

		alias := signUp.Alias.Text
		password := []byte(signUp.Password.Text)
		confirm := []byte(signUp.Confirm.Text)

		err := aliasgo.ValidateAlias(alias)
		if err != nil {
			f.ShowError(err)
			return
		}

		// TODO check Alias is Unique

		if len(password) < cryptogo.MIN_PASSWORD {
			f.ShowError(cryptogo.ErrPasswordTooShort{Size: len(password), Min: cryptogo.MIN_PASSWORD})
			return
		}
		if !bytes.Equal(password, confirm) {
			f.ShowError(cryptogo.ErrPasswordsDoNotMatch{})
			return
		}
		f.NewAccount(client, alias, password, func(account bcgo.Account) {
			if c := callback; c != nil {
				c(account)
			}
			for _, c := range f.onSignedUp {
				c(account)
			}
		})
	}
	signUp.Alias.OnSubmitted = func(string) {
		f.window.Canvas().Focus(signUp.Password)
	}
	signUp.Password.OnSubmitted = func(string) {
		f.window.Canvas().Focus(signUp.Confirm)
	}
	signUp.Confirm.OnSubmitted = func(string) {
		signUpAction()
	}

	signUp.Alias.Validator = func(alias string) error {
		if err := aliasgo.ValidateAlias(alias); err != nil {
			return err
		}

		// TODO check Alias is Unique

		return nil
	}

	signUp.SignUpButton.OnTapped = signUpAction

	rootDir, err := client.Root()
	if err != nil {
		log.Println(err)
	} else {
		keystore, err := bcgo.KeyDirectory(rootDir)
		if err != nil {
			log.Println(err)
		} else {
			keys, err := cryptogo.ListRSAPrivateKeys(keystore)
			if err != nil {
				log.Println(err)
			} else if len(keys) > 0 {
				signIn.Alias.SetOptions(keys)
				signIn.Alias.SetText(keys[0])
				importKey.Alias.SetText(keys[0])
				signUp.Alias.SetText(keys[0])
			}
		}
	}

	if alias, ok := os.LookupEnv("ALIAS"); ok {
		signIn.Alias.SetText(alias)
		importKey.Alias.SetText(alias)
		signUp.Alias.SetText(alias)
	}

	if pwd, ok := os.LookupEnv("PASSWORD"); ok {
		signIn.Password.SetText(pwd)
	}

	if signIn.Alias.Text == "" {
		// Make accordion show sign up as open instead of sign in
		accordion.Open(2)
	}

	// Show Access Dialog
	d.Show()
	d.Resize(ui.DialogSize)
}

func (f *bcFyne) ShowAccount(client bcclientgo.BCClient) {
	account, err := f.Account(client)
	if err != nil {
		f.ShowError(err)
		return
	}
	form, err := identityView(account)
	if err != nil {
		f.ShowError(err)
		return
	}
	contents := container.NewVBox()
	if !bcgo.IsLive() {
		contents.Add(ui.NewTestModeSign())
	}
	contents.Add(form)

	d := dialog.NewCustom("Account", "OK", contents, f.window)
	contents.Add(widget.NewButton("Export Keys", func() {
		f.ExportKeys(client, account)
	}))
	contents.Add(widget.NewButton("Delete Keys", func() {
		d.Hide()
		f.DeleteKeys(client, account)
	}))
	contents.Add(widget.NewButton("Sign Out", func() {
		d.Hide()
		f.SignOut(client)
	}))
	d.Show()
	d.Resize(ui.DialogSize)
}

func (f *bcFyne) DeleteKeys(client bcclientgo.BCClient, account bcgo.Account) {
	f.ShowError(fmt.Errorf("Not yet implemented: %s", "BCFyne.DeleteKeys"))
}

func (f *bcFyne) ExportKeys(client bcclientgo.BCClient, account bcgo.Account) {
	alias := account.Alias()
	authentication := accountui.NewAuthentication(alias)
	authenticateAction := func() {

		host := bcgo.BCWebsite()

		// Show Progress Dialog
		progress := dialog.NewProgress("Exporting Keys", fmt.Sprintf("Exporting %s to %s", alias, host), f.window)
		progress.Show()

		var (
			access string
			err    error
		)

		password := []byte(authentication.Password.Text)
		if len(password) < cryptogo.MIN_PASSWORD {
			err = cryptogo.ErrPasswordTooShort{Size: len(password), Min: cryptogo.MIN_PASSWORD}
		} else {
			access, err = client.ExportKeys(host, alias, password)
		}

		// Hide Progress Dialog
		progress.Hide()

		if err != nil {
			f.ShowError(err)
			return
		}

		form := widget.NewForm(
			widget.NewFormItem("Alias", widget.NewLabel(alias)),
			widget.NewFormItem("Access Code", container.NewHBox(
				widget.NewLabel(access),
				widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
					f.window.Clipboard().SetContent(access)
					dialog.ShowInformation("Copied", "Access code copied to clipboard", f.window)
				}),
			)),
		)
		contents := container.NewVBox()
		if !bcgo.IsLive() {
			contents.Add(ui.NewTestModeSign())
		}
		contents.Add(form)
		d := dialog.NewCustom("Keys Exported", "OK", contents, f.window)
		d.Show()
		d.Resize(ui.DialogSize)

		for _, c := range f.onKeysExported {
			c(alias)
		}
	}
	authentication.Password.OnSubmitted = func(string) {
		authenticateAction()
	}
	authentication.AuthenticateButton.OnTapped = authenticateAction

	contents := container.NewVBox()
	if !bcgo.IsLive() {
		contents.Add(ui.NewTestModeSign())
	}
	contents.Add(authentication.CanvasObject())
	d := dialog.NewCustom("Account", "Cancel", contents, f.window)
	d.Show()
	d.Resize(ui.DialogSize)
}

func (f *bcFyne) SignOut(client bcclientgo.BCClient) {
	client.SetRoot("")
	client.SetCache(nil)
	client.SetNetwork(nil)
	client.SetAccount(nil)
	client.SetNode(nil)
	for _, c := range f.onSignedOut {
		c()
	}
}

func (f *bcFyne) ShowError(err error) {
	log.Println("Error:", err)
	debug.PrintStack()
	dialog.ShowError(err, f.window)
}

func (f *bcFyne) ShowURI(client bcclientgo.BCClient, uri fyne.URI) {
	var view fyne.CanvasObject
	switch u := uri.(type) {
	case storage.AliasURI:
		av := ui.NewAliasView(f, client)
		av.SetURI(u)
		view = av
	case storage.RecordURI:
		rv := ui.NewRecordView(f, client)
		rv.SetURI(u)
		view = rv
	case storage.BlockURI:
		bv := ui.NewBlockView(f, client)
		bv.SetURI(u)
		view = bv
	case storage.ChannelURI:
		hv := ui.NewHeadView(f, client)
		hv.SetURI(u)
		view = hv
	default:
		f.ShowError(fmt.Errorf("Unrecognized URI: %s", uri))
		return
	}

	window := f.app.NewWindow(uri.Name())
	window.SetContent(container.NewVScroll(view))
	window.Resize(ui.WindowSize)
	window.CenterOnScreen()
	window.Show()
}

func (f *bcFyne) ShowIdentity(identity bcgo.Identity) {
	form, err := identityView(identity)
	if err != nil {
		f.ShowError(err)
		return
	}
	contents := container.NewVBox()
	if !bcgo.IsLive() {
		contents.Add(ui.NewTestModeSign())
	}
	contents.Add(form)
	dialog.ShowCustom("Identity", "OK", contents, f.window)
}

func identityView(identity bcgo.Identity) (fyne.CanvasObject, error) {
	bytes, format, err := identity.PublicKey()
	if err != nil {
		return nil, err
	}

	aliasScroller := container.NewHScroll(ui.NewAliasLabel(identity.Alias()))
	bytesScroller := container.NewVScroll(ui.NewKeyLabel(bytes))
	bytesScroller.SetMinSize(fyne.NewSize(0, 10*theme.TextSize())) // Show at least 10 lines
	formatScroller := container.NewVScroll(widget.NewLabel(format.String()))

	return widget.NewForm(
		widget.NewFormItem(
			"Alias",
			aliasScroller,
		),
		widget.NewFormItem(
			"Public Key",
			bytesScroller,
		),
		widget.NewFormItem(
			"Public Key Format",
			formatScroller,
		),
	), nil
}
