import os
from dotenv import load_dotenv

load_dotenv()  # baca ../../.env atau e2e-tests/.env

BASE_URL = os.getenv("E2E_BASE_URL", "http://localhost:8080/v1")
PG_URL   = os.getenv("PG_URL", "postgres://user:myAwEsOm3pa55%40w0rd@localhost:5432/db?sslmode=disable")

# LLM config
LLM_API_KEY  = os.getenv("LLM_API_KEY", "")        # wajib untuk cloud LLM
LLM_BASE_URL = os.getenv("LLM_BASE_URL", "https://api.deepseek.com/v1")
LLM_MODEL    = os.getenv("LLM_MODEL", "deepseek-chat")
LLM_MAX_TOKENS = int(os.getenv("LLM_MAX_TOKENS", "4096"))
