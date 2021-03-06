/*
 * mongodb_ebenchmark - Mongodb grpc proxy benchmark for e-commerce workload (still in dev)
 * Copyright (c) 2020 - Chen, Xidong <chenxidong2009@hotmail.com>
 *
 * All rights reserved.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package main

import (
	"context"
	"github.com/google/uuid"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"github.com/xidongc/mongo_ebenchmark/model/product/productpb"
	product "github.com/xidongc/mongo_ebenchmark/model/product/service"
	sku "github.com/xidongc/mongo_ebenchmark/model/sku/service"
	"github.com/xidongc/mongo_ebenchmark/model/sku/skupb"
	"github.com/xidongc/mongo_ebenchmark/pkg/cfg"
	server "github.com/xidongc/mongo_ebenchmark/pkg/cfg"
	"os"
)

func main() {
	var config server.Config

	parser := flags.NewParser(&config, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
	log.Infof("%+v", config)

	ctx, cancel := context.WithCancel(context.Background())
	proxyConfig := &cfg.ProxyConfig{
		ProxyAddr:    config.ProxyAddr,
		ProxyPort:    config.ProxyPort,
		Secure:       config.Secure,
		RpcTimeout:   config.RpcTimeout,
		BatchSize:    config.BatchSize,
		ReadPref:     config.ReadPref,
		AllowPartial: config.AllowPartial,
	}
	storageClient := *sku.NewClient(proxyConfig, cancel)

	if config.Turbo {
		storageClient.Turbo = true
	}

	defer func() {
		err := storageClient.Close()
		if err != nil {
			log.Error("error")
		}
	}()

	amplifyOptions := &cfg.AmplifyOptions{
		Connections:  config.Connections,
		Concurrency:  config.Concurrency,
		TotalRequest: config.TotalRequest,
		QPS:          config.QPS,
		Timeout:      config.Timeout,
		CPUs:         config.CPUs,
	}

	skuService := &sku.Service{
		Storage:   storageClient,
		Amplifier: amplifyOptions,
	}

	productService := &product.Service{
		Storage:    *product.NewClient(proxyConfig, cancel),
		Amplifier:  amplifyOptions,
		SkuService: skuService,
	}

	productId := uuid.New().String()
	req := &productpb.NewRequest{
		Id:          productId,
		Name:        "mavic",
		Active:      true,
		Attributes:  []string{"drone", "dji"},
		Description: "light weight drone",
		Images: []string{
			"data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wCEAAkGBxMQDxAQDxIVEBUOGBUQEQ8QEBAQEhAQFxUYFxkVGBUaHSggGiYlGxcVIzEhJSkrMDEuFx8zODMsNygtLisBCgoKDg0OGhAQGisdHR0tNS0rLS0rLS0rLS0tLTgyLS0tNS0tKy0tKy4tKy0rKzctLS0tKy0rKysrOC0tLSs3N//AABEIAOEA4QMBIgACEQEDEQH/xAAcAAEAAQUBAQAAAAAAAAAAAAAABgECAwUHBAj/xABEEAACAgECAwQHBAcGBAcAAAABAgADEQQhBRIxBgcTQSIyUWFxgZEUQqGxIzNScoLB0RViksLS4WOisuIWJENEVGRz/8QAFwEBAQEBAAAAAAAAAAAAAAAAAAECA//EABwRAQEBAQADAQEAAAAAAAAAAAABESECEjFRQf/aAAwDAQACEQMRAD8A7jERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQETBq7mRQUra0llXlQ1ggE4LHnYDAG5xvtsDNLf2pNZYW6HXKFJHMmmXUBsHGR4LMcHr0gSGJGR280Q/WPbRjr9o0esoA+b1gfjPXo+2HD7iBVrtM5PRRqauY/wAJOYG7iWVWqwyrBh7VII/CYtXrqqRzXWpUOmbHVBk+WSYHoieevX1MAy2owbcEWKQR7jmeiAiIgIiICIiAiIgImF9VWrqjOoZ/VQuoZvgOpmaAiIgIiICIiAicj70e8G6vULw/hjfpARXa6qrFrnIVaVztsTufaQPIzp/BdG1GmpqsfxHRQLLOnPZ1ZseQLE4HlA9sREBERARKEzScR7W6OgkWahSRsVr5riD7CEBx84G7zPNqeHU27W012fv1o/5iaTR9tdFacC7kP/FrsrH+Ijl/GbrT6yuwA12I4PQo6sD8MGMTWsbsZw/PMNHTWT96qsUn6pgyK9tu7vS+C9ukpKam1kQWfaLMuCdwxsLDoM+W4XfGZ0TeQztt2nFRXTUYe4MHPmtePI+/8hLhrkSdk9eSy4LfZAQVtsp8NT1OHD5P8KjPmPKbTtlwHXaCqqwanUXa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pTx5Y9MxGmVES73OM+Fw/wAJT6eqdUAB35FPMfqQg/ilO6rhfh1u5H6sLQp8y3r2H5kp9DPT2j7EV66+m97bK2oxyheRkODkZVgfOSTgPCU0lIpQs+CWZ7Dlnc7knG3yAwJf4mdbKIiZaIiICUIlYgUCysRAREQEgneRwQalXZrAhFYUB6rHA3bdSjqQdz1yNh7JO5Ge2fi4XwqzaCrq6hWJGccpyM/3tsQOS93vZOiy5/tNj3Y5lCLzVjBwDlixPkOmJGuN9mbdPqLqQrWCtiEcKW5k6qTgdcYz75I+H8Qu0VwblNfiHYW1MpwT1AJH1I+Um3ZjgZ19z6i9yyIQG8msfAOMjZRgjp7cDHlrE38cW/sHUkZXS6hx+0mmvdc+zIXEw2cEvHraa9cft6e1cfVZ9b1VKihEAVVGAqjAA+EozTKvlrQdpeI6M4q1moqA+5a5sT5JblR8hJVwzve1qbX1afVD2gNQ5/iUlf8AlneuXPXf47zBdwnT2frKKn/fprb8xAi3ZztImvTTBtNdU2rSy0hSLK6a1YqGez0SOcg8u2TiYu1ulFlF1OoJR3CrzYKrcFO7KQDuVZthkjpjGCZrptKlShakWtQAoVFCgKNgAB7Jmgc64D2U0V7ahQz3JyIh5mxysw6AcowUCrjb72+ZrD3b6bTWm3WarxaubnXT+GqtaxOeVjkkgn7oG+euNp03V8MptObKkc/tFRzY/e6y3S8KoqbmrqRW6c4Uc+P3jvLqNfwXRu9p1V6lCRy0VNsyIerMPIn2eQ+JA3sRIpERAtIlpSZIgWqsuiICIiAiIgIiICIiAiUzGYFZ4+JaLxlwrmsjcMqo3yKsCCPoffPXmVgfPHbavUVa22i9+fwseEwRawazuCAP9527sjfVZodO+nUIjIDyr5P0bJ6k82dzuZDO+bg3NXTrUG9J8K3/APNuh+R/OYe5/jOFv0bH1f09XwOAw/I/WavYxOV0y2zExoczAH5jPZUmJGl6iXREikREBERAREQEREBERAREQEREBERAREQEREC0iMS6IFspL5TEDX8Y0a6ii2iwZW5SjfPz+RwZwLhOsbQa5OZgXoc1sFOTYnQ4A65U5+c+jCs8a8JpVzYtSBm6uEUMfiZqXEsYOFPzVq+COcBsMMMM+0TZAy0V4l0yqvNK5lhlIGWJQSsBERAREQEREBERAREQEREBERAREQEREBERAREQEREBKYlYgUxGJWICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgf//Z",
			"data:image/jpeg;base64,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",
		},
		Metadata: map[string]string{
			"type":  "electronic",
			"color": "gray",
		},
		Shippable: true,
		Url:       "https://www.wish.com/product/dji-mavic-pro-58ddeb5f927d6d573f1350a1?hide_login_modal=true&share=web",
	}
	if _, err := productService.New(ctx, req); err != nil {
		log.Error(err)
	}

	skus := []*skupb.UpsertRequest{
		{
			Name:              "dji care",
			Currency:          128,
			Active:            true,
			ProductId:         productId,
			Price:             75,
			Metadata:          nil,
			Image:             "",
			SkuLabel:          "dji care",
			PackageDimensions: nil,
			Inventory:         &skupb.Inventory{SkuId: int64(uuid.New().ID()), WarehouseId: int64(uuid.New().ID())},
			Attributes:        nil,
			HasBattery:        false,
			HasLiquid:         false,
			HasSensitive:      false,
			Description:       "dji care insurance",
			Supplier:          "dji",
		},
		{
			Name:      "mavic pro",
			Currency:  128,
			Active:    true,
			ProductId: productId,
			Price:     1000,
			Metadata:  nil,
			Image:     "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wCEAAkGBxMQDxAQDxIVEBUOGBUQEQ8QEBAQEhAQFxUYFxkVGBUaHSggGiYlGxcVIzEhJSkrMDEuFx8zODMsNygtLisBCgoKDg0OGhAQGisdHR0tNS0rLS0rLS0rLS0tLTgyLS0tNS0tKy0tKy4tKy0rKzctLS0tKy0rKysrOC0tLSs3N//AABEIAOEA4QMBIgACEQEDEQH/xAAcAAEAAQUBAQAAAAAAAAAAAAAABgECAwUHBAj/xABEEAACAgECAwQHBAcGBAcAAAABAgADEQQhBRIxBgcTQSIyUWFxgZEUQqGxIzNScoLB0RViksLS4WOisuIWJENEVGRz/8QAFwEBAQEBAAAAAAAAAAAAAAAAAAECA//EABwRAQEBAQADAQEAAAAAAAAAAAABESECEjFRQf/aAAwDAQACEQMRAD8A7jERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQEREBERAREQETBq7mRQUra0llXlQ1ggE4LHnYDAG5xvtsDNLf2pNZYW6HXKFJHMmmXUBsHGR4LMcHr0gSGJGR280Q/WPbRjr9o0esoA+b1gfjPXo+2HD7iBVrtM5PRRqauY/wAJOYG7iWVWqwyrBh7VII/CYtXrqqRzXWpUOmbHVBk+WSYHoieevX1MAy2owbcEWKQR7jmeiAiIgIiICIiAiIgImF9VWrqjOoZ/VQuoZvgOpmaAiIgIiICIiAicj70e8G6vULw/hjfpARXa6qrFrnIVaVztsTufaQPIzp/BdG1GmpqsfxHRQLLOnPZ1ZseQLE4HlA9sREBERARKEzScR7W6OgkWahSRsVr5riD7CEBx84G7zPNqeHU27W012fv1o/5iaTR9tdFacC7kP/FrsrH+Ijl/GbrT6yuwA12I4PQo6sD8MGMTWsbsZw/PMNHTWT96qsUn6pgyK9tu7vS+C9ukpKam1kQWfaLMuCdwxsLDoM+W4XfGZ0TeQztt2nFRXTUYe4MHPmtePI+/8hLhrkSdk9eSy4LfZAQVtsp8NT1OHD5P8KjPmPKbTtlwHXaCqqwanUXa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pTx5Y9MxGmVES73OM+Fw/wAJT6eqdUAB35FPMfqQg/ilO6rhfh1u5H6sLQp8y3r2H5kp9DPT2j7EV66+m97bK2oxyheRkODkZVgfOSTgPCU0lIpQs+CWZ7Dlnc7knG3yAwJf4mdbKIiZaIiICUIlYgUCysRAREQEgneRwQalXZrAhFYUB6rHA3bdSjqQdz1yNh7JO5Ge2fi4XwqzaCrq6hWJGccpyM/3tsQOS93vZOiy5/tNj3Y5lCLzVjBwDlixPkOmJGuN9mbdPqLqQrWCtiEcKW5k6qTgdcYz75I+H8Qu0VwblNfiHYW1MpwT1AJH1I+Um3ZjgZ19z6i9yyIQG8msfAOMjZRgjp7cDHlrE38cW/sHUkZXS6hx+0mmvdc+zIXEw2cEvHraa9cft6e1cfVZ9b1VKihEAVVGAqjAA+EozTKvlrQdpeI6M4q1moqA+5a5sT5JblR8hJVwzve1qbX1afVD2gNQ5/iUlf8AlneuXPXf47zBdwnT2frKKn/fprb8xAi3ZztImvTTBtNdU2rSy0hSLK6a1YqGez0SOcg8u2TiYu1ulFlF1OoJR3CrzYKrcFO7KQDuVZthkjpjGCZrptKlShakWtQAoVFCgKNgAB7Jmgc64D2U0V7ahQz3JyIh5mxysw6AcowUCrjb72+ZrD3b6bTWm3WarxaubnXT+GqtaxOeVjkkgn7oG+euNp03V8MptObKkc/tFRzY/e6y3S8KoqbmrqRW6c4Uc+P3jvLqNfwXRu9p1V6lCRy0VNsyIerMPIn2eQ+JA3sRIpERAtIlpSZIgWqsuiICIiAiIgIiICIiAiUzGYFZ4+JaLxlwrmsjcMqo3yKsCCPoffPXmVgfPHbavUVa22i9+fwseEwRawazuCAP9527sjfVZodO+nUIjIDyr5P0bJ6k82dzuZDO+bg3NXTrUG9J8K3/APNuh+R/OYe5/jOFv0bH1f09XwOAw/I/WavYxOV0y2zExoczAH5jPZUmJGl6iXREikREBERAREQEREBERAREQEREBERAREQEREC0iMS6IFspL5TEDX8Y0a6ii2iwZW5SjfPz+RwZwLhOsbQa5OZgXoc1sFOTYnQ4A65U5+c+jCs8a8JpVzYtSBm6uEUMfiZqXEsYOFPzVq+COcBsMMMM+0TZAy0V4l0yqvNK5lhlIGWJQSsBERAREQEREBERAREQEREBERAREQEREBERAREQEREBKYlYgUxGJWICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgIiICIiAiIgf//Z",
			SkuLabel:  "mavic pro",
			PackageDimensions: &skupb.PackageDimensions{
				Height: 5,
				Weight: 1029,
				Width:  50,
				Length: 100,
			},
			Inventory: &skupb.Inventory{
				SkuId:       int64(uuid.New().ID()),
				WarehouseId: int64(uuid.New().ID()),
			},
			Attributes:   map[string]string{"type": "drone"},
			HasBattery:   true,
			HasLiquid:    false,
			HasSensitive: false,
			Description:  "dji mavic drone",
			Supplier:     "dji",
		},
	}

	for _, request := range skus {
		_, err := skuService.New(ctx, request)
		if err != nil {
			log.Error(err)
		}
	}

	p, err := productService.Get(ctx, &productpb.GetRequest{
		Id: productId,
	})
	if err != nil {
		log.Error(err)
	}
	log.Info(p)

}
