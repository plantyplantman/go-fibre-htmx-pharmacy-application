package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	NEON        string
	TEST_NEON   string
	OPENAI      string
	BIGCOMMERCE string
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	NEON = os.Getenv("NEON_CONNECTION_STRING")
	TEST_NEON = os.Getenv("TEST_NEON_CONNECTION_STRING")
	if NEON == "" && TEST_NEON == "" {
		panic("NEITHER NEON_CONNECTION_STRING NOR TEST_NEON_CONNECTION_STRING IS SET")
	}
	OPENAI = os.Getenv("WORK_OPENAI_API_KEY")
	if OPENAI == "" {
		panic("OPENAI_API_KEY not set")
	}
	BIGCOMMERCE = os.Getenv("BIGCOMMERCE_ACCESS_TOKEN")
	if BIGCOMMERCE == "" {
		panic("BIGCOMMERCE_ACCESS_TOKEN not set")
	}
}
