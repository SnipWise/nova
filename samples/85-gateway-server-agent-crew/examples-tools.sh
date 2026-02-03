#!/bin/bash
# ============================================================
# Exemples d'utilisation des outils avec le gateway server
# Mode Passthrough - Le client gère l'exécution des outils
# ============================================================

BASE_URL="http://localhost:8080"
CONTENT_TYPE="Content-Type: application/json"

echo "============================================"
echo " Gateway Server - Exemples avec outils"
echo " Mode: Passthrough (client-side)"
echo " Target: ${BASE_URL}"
echo "============================================"
echo ""

# ----------------------------------------------------------
# Exemple 1 : Requête simple sans outils
# ----------------------------------------------------------
echo "--- 1. Requête simple sans outils ---"
curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is 2 + 2?"}
    ],
    "stream": false
  }' | jq .
echo ""
echo ""

# ----------------------------------------------------------
# Exemple 2 : Requête avec définition d'outils (le LLM peut les appeler)
# ----------------------------------------------------------
echo "--- 2. Requête avec outils disponibles ---"
echo "Le client déclare les outils disponibles."
echo "Si le LLM décide de les utiliser, il renverra finish_reason: 'tool_calls'"
echo ""

curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "What is the current time?"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_current_time",
          "description": "Get the current time in HH:MM:SS format",
          "parameters": {
            "type": "object",
            "properties": {}
          }
        }
      }
    ],
    "stream": false
  }' | jq .
echo ""
echo "Note: Si le LLM appelle l'outil, la réponse contiendra 'tool_calls'"
echo ""

# ----------------------------------------------------------
# Exemple 3 : Outils multiples disponibles
# ----------------------------------------------------------
echo "--- 3. Plusieurs outils disponibles ---"
echo "Le client peut déclarer plusieurs outils en même temps"
echo ""

curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Calculate 15 + 27"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "calculate_sum",
          "description": "Calculate the sum of two numbers",
          "parameters": {
            "type": "object",
            "properties": {
              "a": {
                "type": "number",
                "description": "First number"
              },
              "b": {
                "type": "number",
                "description": "Second number"
              }
            },
            "required": ["a", "b"]
          }
        }
      },
      {
        "type": "function",
        "function": {
          "name": "calculate_product",
          "description": "Calculate the product of two numbers",
          "parameters": {
            "type": "object",
            "properties": {
              "a": {"type": "number"},
              "b": {"type": "number"}
            },
            "required": ["a", "b"]
          }
        }
      }
    ],
    "stream": false
  }' | jq .
echo ""
echo ""

# ----------------------------------------------------------
# Exemple 4 : Simulation d'un cycle complet tool call -> result -> response
# ----------------------------------------------------------
echo "--- 4. Cycle complet : Tool Call + Result ---"
echo "Étape 1 : Requête initiale avec outil"
echo ""

RESPONSE=$(curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Read the file config.json"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "read_file",
          "description": "Read the contents of a file",
          "parameters": {
            "type": "object",
            "properties": {
              "path": {
                "type": "string",
                "description": "Path to the file"
              }
            },
            "required": ["path"]
          }
        }
      }
    ],
    "stream": false
  }')

echo "$RESPONSE" | jq .
echo ""

# Vérifier si le LLM a appelé un outil
FINISH_REASON=$(echo "$RESPONSE" | jq -r '.choices[0].finish_reason')

if [ "$FINISH_REASON" = "tool_calls" ]; then
    echo "✅ Le LLM a appelé un outil !"
    echo "Dans une vraie application, le client (qwen-code) exécuterait maintenant l'outil"
    echo "et renverrait le résultat dans une nouvelle requête."
    echo ""

    echo "Étape 2 : Le client exécute l'outil et renvoie le résultat"
    echo "(Simulation - en réalité le client lit vraiment le fichier)"
    echo ""

    curl -s "${BASE_URL}/v1/chat/completions" \
      -H "${CONTENT_TYPE}" \
      -d '{
        "model": "crew",
        "messages": [
          {"role": "user", "content": "Read the file config.json"},
          {
            "role": "assistant",
            "content": null,
            "tool_calls": [
              {
                "id": "call_123",
                "type": "function",
                "function": {
                  "name": "read_file",
                  "arguments": "{\"path\":\"config.json\"}"
                }
              }
            ]
          },
          {
            "role": "tool",
            "content": "{\"content\": \"{ \\\"port\\\": 8080, \\\"host\\\": \\\"localhost\\\" }\"}",
            "tool_call_id": "call_123"
          }
        ],
        "tools": [
          {
            "type": "function",
            "function": {
              "name": "read_file",
              "description": "Read the contents of a file",
              "parameters": {
                "type": "object",
                "properties": {
                  "path": {"type": "string"}
                },
                "required": ["path"]
              }
            }
          }
        ],
        "stream": false
      }' | jq .
    echo ""
else
    echo "ℹ️  Le LLM n a pas appelé d outil (finish_reason: $FINISH_REASON)"
    echo "Cela peut arriver si le LLM décide qu il peut répondre sans outil"
fi
echo ""

# ----------------------------------------------------------
# Exemple 5 : Test avec streaming
# ----------------------------------------------------------
echo "--- 5. Requête avec streaming et outils ---"
echo "En mode streaming, les tool_calls sont envoyés par chunks"
echo ""

curl -sN "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Get the weather for Paris"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get weather for a city",
          "parameters": {
            "type": "object",
            "properties": {
              "city": {"type": "string"}
            },
            "required": ["city"]
          }
        }
      }
    ],
    "stream": true
  }'
echo ""
echo ""

# ----------------------------------------------------------
# Exemple 6 : Format content array (compatible qwen-code)
# ----------------------------------------------------------
echo "--- 6. Format content en array (qwen-code) ---"
echo "Le gateway supporte content: [\"text\"] en plus de content: \"text\""
echo ""

curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {
        "role": "user",
        "content": ["Write", "a", "hello", "world", "in", "Go"]
      }
    ],
    "stream": false
  }' | jq .
echo ""
echo ""

# ----------------------------------------------------------
# Exemple 7 : Routage automatique vers l'agent coder
# ----------------------------------------------------------
echo "--- 7. Routage automatique avec outils ---"
echo "Les outils fonctionnent avec tous les agents (coder, thinker, generic)"
echo ""

curl -s "${BASE_URL}/v1/chat/completions" \
  -H "${CONTENT_TYPE}" \
  -d '{
    "model": "crew",
    "messages": [
      {"role": "user", "content": "Write a Go function to reverse a string"}
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "execute_code",
          "description": "Execute code and return the result",
          "parameters": {
            "type": "object",
            "properties": {
              "language": {"type": "string"},
              "code": {"type": "string"}
            },
            "required": ["language", "code"]
          }
        }
      }
    ],
    "stream": false
  }' | jq .
echo ""
echo ""

echo "============================================"
echo " Tous les exemples terminés !"
echo "============================================"
echo ""
echo "💡 Conseils :"
echo "  - En mode Passthrough, les outils sont déclarés par le CLIENT"
echo "  - Le gateway transmet les tool_calls au client"
echo "  - Le client exécute les outils et renvoie les résultats"
echo "  - Le LLM génère la réponse finale basée sur les résultats"
echo ""
echo "📖 Voir README-tools.md pour plus de détails"
