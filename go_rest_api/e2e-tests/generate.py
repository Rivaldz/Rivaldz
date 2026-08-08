import os
import sys
import subprocess
import requests
from prompts.builder import build_prompt
from config import LLM_API_KEY, LLM_BASE_URL, LLM_MODEL, LLM_MAX_TOKENS

def main():
    # Read system prompt
    system_prompt_path = os.path.join(os.path.dirname(__file__), "prompts", "system.txt")
    with open(system_prompt_path, "r") as f:
        system_content = f.read()

    # Build user prompt context
    user_content = build_prompt()

    print(f"Generating E2E tests using {LLM_MODEL} at {LLM_BASE_URL}...")

    # Call LLM API (OpenAI compatible API)
    headers = {
        "Content-Type": "application/json"
    }
    if LLM_API_KEY:
        headers["Authorization"] = f"Bearer {LLM_API_KEY}"

    payload = {
        "model": LLM_MODEL,
        "messages": [
            {"role": "system", "content": system_content},
            {"role": "user", "content": user_content}
        ],
        "temperature": 0.3,
        "max_tokens": LLM_MAX_TOKENS
    }

    try:
        response = requests.post(f"{LLM_BASE_URL.rstrip('/')}/chat/completions", json=payload, headers=headers)
        response.raise_for_status()
        data = response.json()
        llm_output = data["choices"][0]["message"]["content"]
    except Exception as e:
        print(f"Error calling LLM API: {e}")
        if 'response' in locals():
            print(response.text)
        sys.exit(1)

    # Parse response
    code = llm_output
    if "```python" in code:
        code = code.split("```python")[1].split("```")[0].strip()
    elif "```" in code:
        parts = code.split("```")
        if len(parts) >= 3:
            code = parts[1].strip()

    # Save to file
    out_file = os.path.join(os.path.dirname(__file__), "generated", "test_blog_v1.py")
    with open(out_file, "w") as f:
        f.write(code)

    print(f"Test generated and saved to {out_file}")

    # Validate syntax with pytest
    print("Validating syntax...")
    result = subprocess.run(["pytest", "e2e-tests/generated/", "--collect-only"])
    if result.returncode == 0:
        print("Syntax validation passed!")
    else:
        print("Warning: Generated code has syntax errors or pytest warnings.")
        
if __name__ == "__main__":
    main()
