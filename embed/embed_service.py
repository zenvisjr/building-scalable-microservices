from flask import Flask, request, jsonify
import openai
import os
from dotenv import load_dotenv

load_dotenv()

app = Flask(__name__)
MODEL = "text-embedding-3-small"

@app.route("/embed", methods=["POST"])
def embed():
    data = request.get_json()
    text = f"{data.get('name', '')} {data.get('description', '')}".strip()

    if not text:
        return jsonify({"error": "Empty input"}), 400

    try:
        client = openai.OpenAI(api_key=os.getenv("OPENAI_API_KEY"))

        res = client.embeddings.create(
            model=MODEL,
            input=text
        )
        embedding = res.data[0].embedding
        return jsonify({"embedding": embedding})
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5005)
