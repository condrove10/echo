package docker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/condrove10/echo/internal/api/adguard"
	"github.com/condrove10/echo/internal/echo"
	"github.com/condrove10/echo/internal/utils"
	"github.com/condrove10/echo/pkg/dns"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func Startup(ctx context.Context) error {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if err := handleContainerDnsRewrite(&container); err != nil {
			return err
		}
	}

	return nil
}

func Monitor(ctx context.Context) {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Create a channel to receive Docker events
	eventsChan, errorsChan := client.Events(ctx, events.ListOptions{})

	// Start listening for events
	for {
		select {
		case event := <-eventsChan:
			if event.Type == events.ContainerEventType && event.Action == "start" {
				handleContainerEvent(ctx, client, event)
			}
		case err := <-errorsChan:
			log.Printf("Error receiving event: %v", err)
			return
		}
	}
}

func handleContainerEvent(ctx context.Context, client *client.Client, event events.Message) error {
	containerID := event.Actor.ID

	// Create a filter to get only the specific container
	filterArgs := filters.NewArgs()
	filterArgs.Add("id", containerID)

	// List containers with the filter
	containers, err := client.ContainerList(ctx, container.ListOptions{
		Filters: filterArgs,
		All:     true,
	})
	if err != nil {
		return err
	}

	for _, container := range containers {
		if err := handleContainerDnsRewrite(&container); err != nil {
			return err
		}
	}

	return nil
}

func handleContainerDnsRewrite(container *types.Container) error {
	config, err := echo.Decode(container.Labels)
	if err != nil {
		return nil
	}

	for name, service := range config.Services {
		network := container.NetworkSettings.Networks[service.Network]
		if network == nil {
			return fmt.Errorf("couldn't find network '%s' nomination in container '%s'", service.Network, container.ID)
		}

		switch strings.ToLower(service.Api.Name) {
		case "adguard":
			client, err := adguard.New(
				adguard.WithContext(context.TODO()),
				adguard.WithUrl(service.Api.Client.Address),
				adguard.WithAuthHeader(service.Api.Client.Authorization),
				adguard.WithRetries(5),
				adguard.WithRetryDelay(time.Millisecond*50),
				adguard.WithInsecureTls(service.Api.Client.Tls.Skip),
			)

			if err != nil {
				return err
			}

			if service.Enable {
				if err := dns.AppendDnsRecord(client, dns.Record{
					Name:    utils.JoinStrings(".", name, service.Hostname),
					Content: network.IPAddress,
				}); err != nil {
					return err
				}
			} else {
				if err := dns.RemoveDnsRecord(client, dns.Record{
					Name:    utils.JoinStrings(".", name, service.Hostname),
					Content: network.IPAddress,
				}); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("couldn't match API name '%s' of service '%s'", service.Api.Name, name)
		}
	}

	return nil
}
