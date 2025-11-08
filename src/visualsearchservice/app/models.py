"""
Data models for Visual Search Service
"""

from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime


class SimilarProduct(BaseModel):
    """Model for a similar product result"""
    product_id: str = Field(..., description="Product ID")
    similarity_score: float = Field(..., ge=0.0, le=1.0, description="Similarity score (0-1)")
    product_name: Optional[str] = Field(None, description="Product name")
    product_image_url: Optional[str] = Field(None, description="Product image URL")
    price: Optional[float] = Field(None, description="Product price")


class VisualSearchResponse(BaseModel):
    """Response model for visual search"""
    query_image: str = Field(..., description="Uploaded image filename")
    results: List[SimilarProduct] = Field(..., description="List of similar products")
    total_results: int = Field(..., description="Total number of results")
    timestamp: datetime = Field(..., description="Search timestamp")


class HealthCheckResponse(BaseModel):
    """Health check response"""
    status: str = Field(..., description="Service status")
    service: str = Field(..., description="Service name")
    timestamp: datetime = Field(..., description="Current timestamp")
    index_size: Optional[int] = Field(None, description="Number of indexed products")


class IndexProductRequest(BaseModel):
    """Request to index a product"""
    product_id: str
    product_name: Optional[str] = None
    image_url: Optional[str] = None
