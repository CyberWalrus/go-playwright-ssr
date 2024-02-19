package main

import (
	"net/http"
	"strings"

	"compress/gzip"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"

	"github.com/labstack/echo/v4"
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("Server Running https://localhost:3000")

	server := echo.New()

	server.GET("/", pageHandle)
	error := server.Start(":3000")

	if error != nil {
		log.Error().Err(error).Msg("")
	}

}

func getPageHandle() string {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("could not launch playwright: %v")
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatal().Err(err).Msg("could not launch Chromium: %v")
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatal().Err(err).Msg("could not create page: %v")
	}
	if _, err = page.Goto("https://cyberwalrus.github.io/novice-learning-prototype/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		log.Fatal().Err(err).Msg("could not goto: %v")
	}
	content, err := page.Content()
	if err != nil {
		log.Fatal().Err(err).Msg("could not create screenshot: %v")
	}

	if err = browser.Close(); err != nil {
		log.Fatal().Err(err).Msg("could not close browser: %v")
	}
	if err = pw.Stop(); err != nil {
		log.Fatal().Err(err).Msg("could not stop Playwright: %v")
	}

	return content
}

func HtmlMinify(value string) string {
	m := minify.New()

	m.AddFunc("text/html", html.Minify)
	out, err := m.String("text/html", value)
	if err != nil {
		panic(err)
	}
	return out
}

func pageHandle(ctx echo.Context) error {
	content := getPageHandle()

	if strings.Contains(ctx.Request().Header.Get("Accept-Encoding"), "gzip") {
		ctx.Response().Header().Set(echo.HeaderContentEncoding, "gzip")

		gw := gzip.NewWriter(ctx.Response().Writer)
		defer gw.Close()

		ctx.Response().WriteHeader(http.StatusOK)
		if _, err := gw.Write([]byte(HtmlMinify(content))); err != nil {
			return err
		}
	} else {
		return ctx.HTML(http.StatusOK, HtmlMinify(content))
	}

	return nil
}
