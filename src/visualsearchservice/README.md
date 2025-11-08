# Visual Search Service

Image-based product search using deep learning and vector similarity.

## Features

- 🔍 **Image-based Search**: Upload an image to find similar products
- 🧠 **Deep Learning**: Uses MobileNetV2 for feature extraction
- ⚡ **Fast Similarity Search**: FAISS for efficient vector search
- 📊 **Scalable**: Can index thousands of products
- 🎯 **Accurate**: Returns products ranked by visual similarity

## How It Works

1. **Feature Extraction**:
   - Uses pre-trained MobileNetV2 (trained on ImageNet)
   - Extracts 1280-dimensional feature vectors from images
   - Features capture visual characteristics (colors, shapes, textures)

2. **Vector Indexing**:
   - FAISS (Facebook AI Similarity Search) indexes all product features
   - Enables fast nearest-neighbor search
   - Supports millions of products with minimal latency

3. **Similarity Search**:
   - Computes L2 distance between query and indexed features
   - Returns top-k most similar products
   - Converts distances to similarity scores (0-1)

## API Endpoints

### Search for Similar Products

```bash
POST /search
```

Upload an image to find similar products.

**Request:**
```bash
curl -X POST "http://localhost:8090/search?top_k=5&threshold=0.7" \
  -F "image=@product_image.jpg"
```

**Parameters:**
- `image` (file): Image file (JPEG, PNG)
- `top_k` (int): Number of results (1-20, default: 5)
- `threshold` (float): Minimum similarity (0-1, default: 0.7)

**Response:**
```json
{
  "query_image": "product_image.jpg",
  "results": [
    {
      "product_id": "OLJCESPC7Z",
      "similarity_score": 0.95,
      "product_name": "Sunglasses",
      "product_image_url": "/static/img/products/sunglasses.jpg",
      "price": 19.99
    }
  ],
  "total_results": 5,
  "timestamp": "2024-01-15T10:30:00"
}
```

### Index a Product

```bash
POST /index/product
```

Add a product image to the search index.

**Request:**
```bash
curl -X POST "http://localhost:8090/index/product?product_id=NEW_PRODUCT_123" \
  -F "image=@new_product.jpg"
```

### Batch Index Products

```bash
POST /index/batch
```

Index all products from the product catalog.

### Health Check

```bash
GET /health
```

Check service health and get statistics.

## Running Locally

### With Docker

```bash
# Build image
docker build -t visualsearchservice .

# Run container
docker run -p 8090:8090 visualsearchservice
```

### With Python

```bash
# Install dependencies
pip install -r requirements.txt

# Run service
python -m app.main

# Or with uvicorn
uvicorn app.main:app --reload --port 8090
```

## Usage Examples

### Python Client

```python
import requests

# Search for similar products
with open('my_image.jpg', 'rb') as f:
    response = requests.post(
        'http://localhost:8090/search',
        files={'image': f},
        params={'top_k': 5, 'threshold': 0.7}
    )

results = response.json()
for product in results['results']:
    print(f"{product['product_name']}: {product['similarity_score']:.2f}")
```

### JavaScript (Frontend)

```javascript
async function visualSearch(imageFile) {
    const formData = new FormData();
    formData.append('image', imageFile);

    const response = await fetch(
        'http://localhost:8090/search?top_k=5',
        {
            method: 'POST',
            body: formData
        }
    );

    const results = await response.json();
    return results.results;
}
```

### cURL

```bash
# Search
curl -X POST "http://localhost:8090/search?top_k=3" \
  -F "image=@sunglasses.jpg"

# Index new product
curl -X POST "http://localhost:8090/index/product?product_id=ABC123" \
  -F "image=@new_sunglasses.jpg"

# Health check
curl http://localhost:8090/health
```

## Integration with Frontend

Add image upload to product pages:

```html
<!-- Image Search -->
<form id="visual-search-form">
    <input type="file" id="image-upload" accept="image/*">
    <button type="submit">Search by Image</button>
</form>

<div id="results"></div>

<script>
document.getElementById('visual-search-form').addEventListener('submit', async (e) => {
    e.preventDefault();

    const imageFile = document.getElementById('image-upload').files[0];
    const formData = new FormData();
    formData.append('image', imageFile);

    const response = await fetch('http://localhost:8090/search?top_k=6', {
        method: 'POST',
        body: formData
    });

    const data = await response.json();

    // Display results
    const resultsDiv = document.getElementById('results');
    resultsDiv.innerHTML = data.results.map(product => `
        <div class="product-card">
            <img src="${product.product_image_url}" alt="${product.product_name}">
            <h3>${product.product_name}</h3>
            <p>Similarity: ${(product.similarity_score * 100).toFixed(0)}%</p>
            <p>Price: $${product.price}</p>
        </div>
    `).join('');
});
</script>
```

## Performance

- **Feature Extraction**: ~100ms per image (CPU)
- **Search**: <10ms for 10,000 products
- **Memory**: ~1GB for 10,000 products (with MobileNetV2)

## Configuration

Environment variables:

```bash
# Port
PORT=8090

# Model selection
MODEL_NAME=MobileNetV2

# Index persistence
INDEX_PATH=data/faiss_index.bin
METADATA_PATH=data/metadata.pkl
```

## Advanced Features

### Custom Similarity Threshold

Adjust the threshold based on your use case:
- `threshold=0.9`: Very strict (only near-exact matches)
- `threshold=0.7`: Balanced (default)
- `threshold=0.5`: Lenient (broader matches)

### Batch Processing

For indexing many products at once:

```python
import glob
import requests

for image_path in glob.glob('products/*.jpg'):
    product_id = image_path.split('/')[-1].replace('.jpg', '')

    with open(image_path, 'rb') as f:
        requests.post(
            'http://localhost:8090/index/product',
            files={'image': f},
            params={'product_id': product_id}
        )
```

## Troubleshooting

### Model Download

On first run, TensorFlow will download MobileNetV2 weights (~14MB). This may take a few minutes.

### Memory Issues

If running out of memory:
1. Use MobileNetV2 instead of larger models
2. Reduce batch size
3. Use GPU for feature extraction

### Slow Search

If search is slow:
1. Ensure FAISS is using CPU BLAS (installed by default)
2. Consider using GPU FAISS for large indices
3. Pre-normalize feature vectors

## Future Enhancements

- [ ] GPU support for faster feature extraction
- [ ] Support for video search
- [ ] Multi-modal search (text + image)
- [ ] Real-time product catalog sync
- [ ] A/B testing different models
- [ ] Product category filtering
- [ ] Color-based filtering

## References

- [MobileNetV2 Paper](https://arxiv.org/abs/1801.04381)
- [FAISS Documentation](https://github.com/facebookresearch/faiss)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
