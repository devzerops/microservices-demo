# Review & Rating Service

A microservice for managing product reviews and ratings in the Online Boutique e-commerce platform.

## Features

- **Product Reviews**: Create, read, update, and delete product reviews
- **Star Ratings**: 1-5 star rating system
- **User Reactions**: Mark reviews as helpful or report inappropriate content
- **Review Statistics**: Automatic calculation of average ratings and rating breakdowns
- **Advanced Filtering**: Filter reviews by minimum rating
- **Flexible Sorting**: Sort by rating, date, or helpfulness
- **Verified Purchases**: Mark reviews from verified purchasers
- **Image Support**: Attach images to reviews (optional)

## API Endpoints

### Review Operations

#### Create Review
```http
POST /reviews
Content-Type: application/json

{
  "product_id": "OLJCESPC7Z",
  "user_id": "user123",
  "username": "John Doe",
  "rating": 5,
  "title": "Amazing product!",
  "content": "This product exceeded all my expectations. Highly recommended!",
  "verified_purchase": true,
  "images": ["https://example.com/image1.jpg"]
}
```

#### Get Review
```http
GET /reviews/{review_id}
```

#### Update Review
```http
PUT /reviews/{review_id}
Content-Type: application/json

{
  "rating": 4,
  "title": "Updated title",
  "content": "Updated content"
}
```

#### Delete Review
```http
DELETE /reviews/{review_id}
```

### Product Reviews

#### Get All Reviews for a Product
```http
GET /products/{product_id}/reviews?min_rating=4&sort_by=helpful

Query Parameters:
- min_rating: Filter by minimum rating (1-5)
- sort_by: Sort order (recent, oldest, rating_desc, rating_asc, helpful)
```

#### Get Product Statistics
```http
GET /products/{product_id}/stats

Response:
{
  "product_id": "OLJCESPC7Z",
  "total_reviews": 42,
  "average_rating": 4.5,
  "rating_breakdown": {
    "1": 2,
    "2": 3,
    "3": 5,
    "4": 12,
    "5": 20
  },
  "last_updated": "2024-11-09T10:30:00Z"
}
```

### Reactions

#### Add Reaction (Helpful/Report)
```http
POST /reviews/{review_id}/reactions
Content-Type: application/json

{
  "user_id": "user456",
  "reaction_type": "helpful"
}
```

#### Remove Reaction
```http
DELETE /reviews/{review_id}/reactions
Content-Type: application/json

{
  "user_id": "user456"
}
```

### Health Check
```http
GET /health
```

## Data Models

### Review
```go
{
  "review_id": "uuid",
  "product_id": "string",
  "user_id": "string",
  "username": "string",
  "rating": 1-5,
  "title": "string",
  "content": "string",
  "verified_purchase": boolean,
  "images": ["url"],
  "helpful_count": integer,
  "report_count": integer,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Review Statistics
```go
{
  "product_id": "string",
  "total_reviews": integer,
  "average_rating": float,
  "rating_breakdown": {
    "1": count,
    "2": count,
    "3": count,
    "4": count,
    "5": count
  },
  "last_updated": "timestamp"
}
```

## Running the Service

### Locally
```bash
cd src/reviewservice
go run .
```

### With Docker
```bash
docker build -t reviewservice .
docker run -p 8096:8096 reviewservice
```

### Environment Variables
- `PORT`: Service port (default: 8096)

## Testing

Run unit tests:
```bash
go test -v
```

Run tests with coverage:
```bash
go test -v -cover
```

## Business Rules

1. **Review Content**:
   - Title is required
   - Content must be at least 10 characters
   - Content must be less than 5000 characters
   - Rating must be between 1-5 stars

2. **User Reactions**:
   - Each user can only have one reaction per review
   - Changing reaction updates the counts appropriately
   - Supported reactions: "helpful", "report"

3. **Statistics**:
   - Automatically updated when reviews are created, updated, or deleted
   - Average rating calculated from all reviews for a product
   - Rating breakdown shows distribution across 1-5 stars

## Integration with Other Services

This service integrates with:
- **Product Catalog**: Reviews reference product IDs
- **User Service**: Reviews reference user IDs
- **API Gateway**: Routes requests to this service
- **Frontend**: Displays reviews and ratings

## Future Enhancements

- [ ] Persistent storage (PostgreSQL/MongoDB)
- [ ] Review moderation workflow
- [ ] Image upload and storage
- [ ] Review voting/ranking algorithm
- [ ] Spam detection
- [ ] Reply to reviews
- [ ] Review templates
- [ ] Export reviews (CSV, JSON)
- [ ] Analytics and insights

## Architecture

The service is built with:
- **Language**: Go 1.21
- **HTTP Framework**: Gorilla Mux
- **Storage**: In-memory (for demo; easily replaceable with DB)
- **API Style**: REST
- **Port**: 8096

## License

Apache License 2.0
