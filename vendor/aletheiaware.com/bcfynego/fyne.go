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
	"aletheiaware.com/bcfynego/ui/account"
	"aletheiaware.com/bcfynego/ui/data"
	"aletheiaware.com/bcgo"
	"aletheiaware.com/cryptogo"
	"bytes"
	"errors"
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
)

type BCFyne struct {
	App            fyne.App
	Window         fyne.Window
	Dialog         dialog.Dialog
	OnKeysExported func(string)
	OnKeysImported func(string)
	OnSignedIn     func(*bcgo.Node)
	OnSignedUp     func(*bcgo.Node)
	OnSignedOut    func()
}

func NewBCFyne(a fyne.App, w fyne.Window) *BCFyne {
	return &BCFyne{
		App:    a,
		Window: w,
	}
}

func (f *BCFyne) ExistingNode(client *bcclientgo.BCClient, alias string, password []byte, callback func(*bcgo.Node)) {
	rootDir, err := client.GetRoot()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Get key store
	keystore, err := bcgo.GetKeyDirectory(rootDir)
	if err != nil {
		f.ShowError(err)
		return
	}
	// Get private key
	key, err := cryptogo.GetRSAPrivateKey(keystore, alias, password)
	if err != nil {
		f.ShowError(err)
		return
	}
	cache, err := client.GetCache()
	if err != nil {
		f.ShowError(err)
		return
	}
	network, err := client.GetNetwork()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Create node
	node := &bcgo.Node{
		Alias:    alias,
		Key:      key,
		Cache:    cache,
		Network:  network,
		Channels: make(map[string]*bcgo.Channel),
	}

	if c := callback; c != nil {
		c(node)
	}
}

func (f *BCFyne) GetNode(client *bcclientgo.BCClient) (*bcgo.Node, error) {
	if client.Node == nil {
		nc := make(chan *bcgo.Node, 1)
		go f.ShowAccessDialog(client, func(n *bcgo.Node) {
			nc <- n
		})
		client.Node = <-nc
	}
	return client.Node, nil
}

func (f *BCFyne) GetLogo() fyne.CanvasObject {
	return &canvas.Image{
		Resource: data.Logo,
		//FillMode: canvas.ImageFillContain,
		FillMode: canvas.ImageFillOriginal,
	}
}

func (f *BCFyne) NewNode(client *bcclientgo.BCClient, alias string, password []byte, callback func(*bcgo.Node)) {
	// Show progress dialog
	progress := dialog.NewProgressInfinite("Creating", "Creating "+alias, f.Window)
	progress.Show()
	defer progress.Hide()

	rootDir, err := client.GetRoot()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Get key store
	keystore, err := bcgo.GetKeyDirectory(rootDir)
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
	cache, err := client.GetCache()
	if err != nil {
		f.ShowError(err)
		return
	}
	network, err := client.GetNetwork()
	if err != nil {
		f.ShowError(err)
		return
	}
	// Create node
	node := &bcgo.Node{
		Alias:    alias,
		Key:      key,
		Cache:    cache,
		Network:  network,
		Channels: make(map[string]*bcgo.Channel),
	}

	{
		// Show Progress Dialog
		progress := dialog.NewProgress("Registering", "Registering "+alias, f.Window)
		progress.Show()
		defer progress.Hide()
		listener := &ui.ProgressMiningListener{Func: progress.SetValue}

		// Register Alias
		if err := aliasgo.Register(node, listener); err != nil {
			f.ShowError(err)
			return
		}
	}

	if c := callback; c != nil {
		c(node)
	}
}

func (f *BCFyne) ShowAccessDialog(client *bcclientgo.BCClient, callback func(*bcgo.Node)) {
	signIn := account.NewSignIn()
	importKey := account.NewImportKey()
	signUp := account.NewSignUp()
	if d := f.Dialog; d != nil {
		d.Hide()
	}
	tos := &widget.Hyperlink{Text: "Terms of Service"}
	tos.SetURLFromString("https://aletheiaware.com/terms-of-service.html")
	pp := &widget.Hyperlink{Text: "Privacy Policy", Alignment: fyne.TextAlignTrailing}
	pp.SetURLFromString("https://aletheiaware.com/privacy-policy.html")
	f.Dialog = dialog.NewCustom("Account Access", "Cancel",
		container.NewVBox(
			widget.NewAccordion(
				&widget.AccordionItem{Title: "Sign In", Detail: signIn.CanvasObject(), Open: true},
				widget.NewAccordionItem("Import Keys", importKey.CanvasObject()),
				widget.NewAccordionItem("Sign Up", signUp.CanvasObject()),
			),
			container.NewGridWithColumns(2, tos, pp),
		),
		f.Window)

	signInAction := func() {
		if d := f.Dialog; d != nil {
			d.Hide()
		}
		alias := signIn.Alias.Text
		password := []byte(signIn.Password.Text)
		if len(password) < cryptogo.MIN_PASSWORD {
			f.ShowError(fmt.Errorf(cryptogo.ERROR_PASSWORD_TOO_SHORT, len(password), cryptogo.MIN_PASSWORD))
			return
		}
		f.ExistingNode(client, alias, password, func(node *bcgo.Node) {
			if c := callback; c != nil {
				c(node)
			}
			if c := f.OnSignedIn; c != nil {
				go c(node)
			}
		})
	}
	signIn.Alias.OnSubmitted = func(string) {
		f.Window.Canvas().Focus(signIn.Password)
	}
	signIn.Password.OnSubmitted = func(string) {
		signInAction()
	}
	signIn.SignInButton.OnTapped = signInAction
	importKeyAction := func() {
		if d := f.Dialog; d != nil {
			d.Hide()
		}

		host := bcgo.GetBCWebsite()
		alias := importKey.Alias.Text
		access := importKey.Access.Text

		// Show Progress Dialog
		progress := dialog.NewProgress("Importing Keys", fmt.Sprintf("Importing %s from %s", alias, host), f.Window)
		progress.Show()

		err := client.ImportKeys(host, alias, access)

		progress.Hide()

		if err != nil {
			f.ShowError(err)
			return
		}

		if c := f.OnKeysImported; c != nil {
			go c(alias)
		}

		authentication := account.NewAuthentication(alias)
		authenticateAction := func() {
			if d := f.Dialog; d != nil {
				d.Hide()
			}
			password := []byte(authentication.Password.Text)
			if len(password) < cryptogo.MIN_PASSWORD {
				f.ShowError(fmt.Errorf(cryptogo.ERROR_PASSWORD_TOO_SHORT, len(password), cryptogo.MIN_PASSWORD))
				return
			}
			f.ExistingNode(client, alias, password, func(node *bcgo.Node) {
				if c := callback; c != nil {
					c(node)
				}
				if c := f.OnSignedIn; c != nil {
					go c(node)
				}
			})
		}
		authentication.Password.OnSubmitted = func(string) {
			authenticateAction()
		}
		authentication.AuthenticateButton.OnTapped = authenticateAction

		// Show Success Dialog
		f.Dialog = dialog.NewCustom("Keys Imported", "Cancel",
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Keys for %s successfully imported from %s.\nAuthenticate to continue", alias, host)),
				authentication.CanvasObject()), f.Window)
		f.Dialog.Show()
		f.Dialog.Resize(ui.DialogSize)
	}
	importKey.Alias.OnSubmitted = func(string) {
		f.Window.Canvas().Focus(importKey.Access)
	}
	importKey.Access.OnSubmitted = func(string) {
		importKeyAction()
	}
	importKey.ImportKeyButton.OnTapped = importKeyAction
	signUpAction := func() {
		if d := f.Dialog; d != nil {
			d.Hide()
		}
		alias := signUp.Alias.Text
		password := []byte(signUp.Password.Text)
		confirm := []byte(signUp.Confirm.Text)

		err := aliasgo.ValidateAlias(alias)
		if err != nil {
			f.ShowError(err)
			return
		}

		if len(password) < cryptogo.MIN_PASSWORD {
			f.ShowError(fmt.Errorf(cryptogo.ERROR_PASSWORD_TOO_SHORT, len(password), cryptogo.MIN_PASSWORD))
			return
		}
		if !bytes.Equal(password, confirm) {
			f.ShowError(errors.New(cryptogo.ERROR_PASSWORDS_DO_NOT_MATCH))
			return
		}
		f.NewNode(client, alias, password, func(node *bcgo.Node) {
			if c := callback; c != nil {
				c(node)
			}
			if c := f.OnSignedUp; c != nil {
				go c(node)
			}
		})
	}
	signUp.Alias.OnSubmitted = func(string) {
		f.Window.Canvas().Focus(signUp.Password)
	}
	signUp.Password.OnSubmitted = func(string) {
		f.Window.Canvas().Focus(signUp.Confirm)
	}
	signUp.Confirm.OnSubmitted = func(string) {
		signUpAction()
	}

	signUp.SignUpButton.OnTapped = signUpAction

	rootDir, err := client.GetRoot()
	if err != nil {
		log.Println(err)
	} else {
		keystore, err := bcgo.GetKeyDirectory(rootDir)
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

	f.Dialog.Show()
	f.Dialog.Resize(ui.DialogSize)
}

func (f *BCFyne) ShowAccount(client *bcclientgo.BCClient) {
	node, err := f.GetNode(client)
	if err != nil {
		f.ShowError(err)
		return
	}
	form, err := nodeView(node)
	if err != nil {
		f.ShowError(err)
		return
	}
	box := container.NewVBox(
		form,
	)
	if d := f.Dialog; d != nil {
		d.Hide()
	}
	f.Dialog = dialog.NewCustom("Account", "OK", box, f.Window)
	box.Add(widget.NewButton("Export Keys", func() {
		f.ExportKeys(client, node)
	}))
	box.Add(widget.NewButton("Delete Keys", func() {
		f.Dialog.Hide()
		f.DeleteKeys(client, node)
	}))
	box.Add(widget.NewButton("Sign Out", func() {
		f.Dialog.Hide()
		f.SignOut(client)
	}))
	f.Dialog.Show()
	f.Dialog.Resize(ui.DialogSize)
}

func (f *BCFyne) DeleteKeys(client *bcclientgo.BCClient, node *bcgo.Node) {
	f.ShowError(fmt.Errorf("Not yet implemented: %s", "BCFyne.DeleteKeys"))
}

func (f *BCFyne) ExportKeys(client *bcclientgo.BCClient, node *bcgo.Node) {
	authentication := account.NewAuthentication(node.Alias)
	authenticateAction := func() {
		if d := f.Dialog; d != nil {
			d.Hide()
		}

		host := bcgo.GetBCWebsite()

		// Show Progress Dialog
		progress := dialog.NewProgress("Exporting Keys", fmt.Sprintf("Exporting %s to %s", node.Alias, host), f.Window)
		progress.Show()

		var (
			access string
			err    error
		)

		password := []byte(authentication.Password.Text)
		if len(password) < cryptogo.MIN_PASSWORD {
			err = fmt.Errorf(cryptogo.ERROR_PASSWORD_TOO_SHORT, len(password), cryptogo.MIN_PASSWORD)
		} else {
			access, err = client.ExportKeys(host, node.Alias, password)
		}

		progress.Hide()

		if err != nil {
			f.ShowError(err)
			return
		}

		form := widget.NewForm(
			widget.NewFormItem("Alias", widget.NewLabel(node.Alias)),
			widget.NewFormItem("Access Code", container.NewHBox(
				widget.NewLabel(access),
				widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
					f.Window.Clipboard().SetContent(access)
					dialog.ShowInformation("Copied", "Access code copied to clipboard", f.Window)
				}),
			)),
		)
		f.Dialog = dialog.NewCustom("Keys Exported", "OK", form, f.Window)
		f.Dialog.Show()
		f.Dialog.Resize(ui.DialogSize)

		if c := f.OnKeysExported; c != nil {
			go c(node.Alias)
		}
	}
	authentication.Password.OnSubmitted = func(string) {
		authenticateAction()
	}
	authentication.AuthenticateButton.OnTapped = authenticateAction
	f.Dialog = dialog.NewCustom("Account", "Cancel", authentication.CanvasObject(), f.Window)
	f.Dialog.Show()
	f.Dialog.Resize(ui.DialogSize)
}

func (f *BCFyne) SignOut(client *bcclientgo.BCClient) {
	client.Root = ""
	client.Cache = nil
	client.Network = nil
	client.Node = nil
	if c := f.OnSignedOut; c != nil {
		go c()
	}
}

func (f *BCFyne) ShowError(err error) {
	log.Println("Error:", err)
	debug.PrintStack()
	if d := f.Dialog; d != nil {
		d.Hide()
	}
	f.Dialog = dialog.NewError(err, f.Window)
	f.Dialog.Show()
}

func (f *BCFyne) ShowURI(client *bcclientgo.BCClient, uri fyne.URI) {
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

	window := f.App.NewWindow(uri.Name())
	window.SetContent(container.NewVScroll(view))
	window.Resize(ui.WindowSize)
	window.CenterOnScreen()
	window.Show()
}

func (f *BCFyne) ShowNode(node *bcgo.Node) {
	form, err := nodeView(node)
	if err != nil {
		f.ShowError(err)
		return
	}
	if d := f.Dialog; d != nil {
		d.Hide()
	}
	f.Dialog = dialog.NewCustom("Node", "OK", form, f.Window)
	f.Dialog.Show()
}

func nodeView(node *bcgo.Node) (fyne.CanvasObject, error) {
	publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(&node.Key.PublicKey)
	if err != nil {
		return nil, err
	}

	aliasScroller := container.NewHScroll(ui.NewAliasLabel(node.Alias))
	publicKeyScroller := container.NewVScroll(ui.NewKeyLabel(publicKeyBytes))
	publicKeyScroller.SetMinSize(fyne.NewSize(0, 10*theme.TextSize())) // Show at least 10 lines

	return widget.NewForm(
		widget.NewFormItem(
			"Alias",
			aliasScroller,
		),
		widget.NewFormItem(
			"Public Key",
			publicKeyScroller,
		),
	), nil
}
