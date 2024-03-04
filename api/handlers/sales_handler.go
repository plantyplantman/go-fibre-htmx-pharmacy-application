package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/pkg/sales"
	"github.com/samber/lo"
)

func GetSales(serviceURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fromQ := c.Query("from")
		toQ := c.Query("to")
		if fromQ == "" || toQ == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": "please specify `from` and `to` query params",
			})
		}

		var storeIDs []int
		storeIDsQ := c.Query("stores")
		if storeIDsQ == "" {
			return c.JSON(fiber.Map{
				"error": "please specify `stores` query param as comma separated list of integers",
			})
		} else {
			ids := strings.Split(storeIDsQ, ",")
			if len(ids) == 0 {
				c.Status(http.StatusBadRequest)
				return c.JSON(fiber.Map{
					"error": "please specify `stores` query param as comma separated list of integers",
				})
			}

			storeIDs = lo.FilterMap(ids, func(item string, _ int) (int, bool) {
				id, err := strconv.Atoi(item)
				if err != nil {
					return 0, false
				}
				return id, true
			})
		}

		log.Println(storeIDs)

		factories := lo.FilterMap(storeIDs, func(storeID int, _ int) (sales.ReportFactory, bool) {
			s, err := sales.NewService(storeID, serviceURL)
			f := sales.NewReportFactory(s)
			return f, err == nil
		})

		if len(factories) == 0 {
			c.Status(http.StatusInternalServerError)
			return c.JSON(fiber.Map{
				"error": "no valid stores found",
			})
		}

		from, err := time.Parse(`2006-01-02T15:04:05`, fromQ)
		if err != nil {
			from, err = time.Parse(`2006-01-02`, fromQ)
			if err != nil {
				c.Status(http.StatusBadRequest)
				return c.JSON(fiber.Map{
					"error": err.Error(),
				})
			}
		}

		to, err := time.Parse(`2006-01-02T15:04:05`, toQ)
		if err != nil {
			to, err = time.Parse(`2006-01-02`, toQ)
			if err != nil {
				c.Status(http.StatusBadRequest)
				return c.JSON(fiber.Map{
					"error": err.Error(),
				})
			}
		}

		intervalQ := c.Query("interval")
		if intervalQ == "" {
			report, err := factories[0].GenerateReport(from, to)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return c.JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.JSON(report)
		}

		interval, err := time.ParseDuration(intervalQ)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		var reports = lo.FilterMap(factories, func(f sales.ReportFactory, _ int) ([]*sales.Report, bool) {
			r, err := f.GenerateReports(from, to, interval)
			return r, err == nil
		})
		return c.JSON(reports)
	}
}
