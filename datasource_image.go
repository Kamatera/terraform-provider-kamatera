package main

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func dataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: DataSourceImageRead,

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"datacenter_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"os": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"code": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func getImageMatchesBy(images map[string]map[string]string, attr string, value string) []string {
	var matchIds []string
	if value != "" {
		for imageId, image := range images {
			if image[attr] == value {
				matchIds = append(matchIds, imageId)
			}
		}
	}
	return matchIds
}

func getAvailableImages(images map[string]map[string]string) string {
	var availableImages []string
	availableImages = append(availableImages, fmt.Sprintf(
		"%-10s %-30s %s", "os", "code", "name",
	))
	for _, image := range images {
		availableImages = append(availableImages, fmt.Sprintf(
			"%-10s %-30s %s",
			"\"" + image["os"] + "\"",
			"\"" + image["code"] + "\"",
			"\"" + image["name"] + "\"",
		))
	}
	return strings.Join(availableImages, "\n")
}

func DataSourceImageRead(d *schema.ResourceData, m interface{}) error {
	provider := m.(*ProviderConfiguration)
	result, e := kamateraRequest(*provider, "GET", fmt.Sprintf("service/server?images=1&datacenter=%s", d.Get("datacenter_id").(string)), nil)
	if e != nil {
		d.SetId("")
		return e
	}
	images := map[string]map[string]string{}
	for _, image := range result.([]interface{}) {
		images[image.(map[string]interface{})["id"].(string)] = map[string]string{
			"os": image.(map[string]interface{})["os"].(string),
			"code": image.(map[string]interface{})["code"].(string),
			"name": image.(map[string]interface{})["name"].(string),
		}
	}
	id := d.Get("id").(string)
	os := d.Get("os").(string)
	code := d.Get("code").(string)
	image, hasImage := images[id]
	osImageIds := getImageMatchesBy(images, "os", os)
	codeImageIds := getImageMatchesBy(images, "code", code)
	if hasImage &&
			(len(osImageIds) == 0   || (len(osImageIds)   == 1 && osImageIds[0]   == image["id"] && image["os"]   == os   )) &&
			(len(codeImageIds) == 0 || (len(codeImageIds) == 1 && codeImageIds[0] == image["id"] && image["code"] == code )) {
		d.SetId(image["id"])
		d.Set("code", image["code"])
		d.Set("os", image["os"])
		return nil
	} else if len(osImageIds) == 1 &&
			(! hasImage || image["os"] == os) &&
			(len(codeImageIds) == 0  || (len(codeImageIds) == 1 && codeImageIds[0] == osImageIds[0])) {
		d.SetId(osImageIds[0])
		d.Set("code", images[osImageIds[0]]["code"])
		d.Set("os", os)
		return nil
	} else if len(codeImageIds) == 1 &&
		(! hasImage || image["code"] == code) {
		d.SetId(codeImageIds[0])
		d.Set("code", code)
		d.Set("os", images[codeImageIds[0]]["os"])
		return nil
	} else {
		d.SetId("")
		d.Set("code", "")
		d.Set("os", "")
		return errors.New(fmt.Sprintf("could not find matching image, available images: \n%s", getAvailableImages(images)))
	}
}
