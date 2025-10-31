from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import uvicorn

app = FastAPI()
model = SentenceTransformer("intfloat/e5-base-v2")

class EmbedRequest(BaseModel):
    text: str

@app.post("/embed")
async def embed(req: EmbedRequest): # type: ignore
    vector = model.encode(req.text).tolist() # type: ignore
    return {"vector": vector} # type: ignore

if __name__ == "__main__":
    print("starting server at: 5005")
    uvicorn.run(app, host="0.0.0.0", port=5005)