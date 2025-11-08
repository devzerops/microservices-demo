# 실험적인 데모 서비스 제안

마이크로서비스 데모 프로젝트에 추가하면 좋을 혁신적이고 실험적인 서비스들

---

## 🤖 AI/ML 기반 서비스

### 1. Visual Search Service (이미지 기반 제품 검색)

**기술 스택:** Python, TensorFlow/PyTorch, OpenCV, FastAPI

**기능:**
- 사용자가 업로드한 이미지로 유사 제품 찾기
- 이미지 내 여러 제품 자동 감지
- 스타일 기반 추천 ("이 스타일과 어울리는 제품")

**구현 예시:**
```python
# visualsearchservice/visual_search.py
import tensorflow as tf
from PIL import Image
import numpy as np

class VisualSearchService:
    def __init__(self):
        self.model = tf.keras.applications.MobileNetV2(weights='imagenet')
        self.feature_extractor = tf.keras.Model(
            inputs=self.model.input,
            outputs=self.model.layers[-2].output
        )

    def extract_features(self, image):
        """이미지에서 특징 벡터 추출"""
        img = Image.open(image).resize((224, 224))
        img_array = np.array(img) / 255.0
        features = self.feature_extractor.predict(np.expand_dims(img_array, axis=0))
        return features

    def find_similar_products(self, image, top_k=5):
        """유사한 제품 찾기"""
        query_features = self.extract_features(image)
        # Vector similarity search (FAISS, Annoy 등 사용)
        similar_products = self.vector_db.search(query_features, k=top_k)
        return similar_products
```

**데모 가치:**
- Computer Vision 실전 활용
- 벡터 데이터베이스 (Milvus, Pinecone) 통합
- ML 모델 서빙 패턴

**마이크로서비스 연동:**
```
User Upload Image
  → Visual Search Service (특징 추출)
  → Vector DB (유사도 검색)
  → Product Catalog (제품 정보)
  → Frontend (결과 표시)
```

---

### 2. Sentiment Analysis Service (리뷰 감정 분석)

**기술 스택:** Python, Transformers (BERT), FastAPI

**기능:**
- 제품 리뷰 자동 감정 분석
- 긍정/부정/중립 분류 및 점수화
- 주요 키워드/토픽 추출
- 가짜 리뷰 탐지

**구현 예시:**
```python
from transformers import pipeline

class SentimentAnalyzer:
    def __init__(self):
        self.sentiment_pipeline = pipeline(
            "sentiment-analysis",
            model="nlptown/bert-base-multilingual-uncased-sentiment"
        )
        self.keyword_extractor = pipeline("ner")

    def analyze_review(self, review_text):
        """리뷰 감정 분석"""
        sentiment = self.sentiment_pipeline(review_text)[0]
        keywords = self.keyword_extractor(review_text)

        return {
            'sentiment': sentiment['label'],
            'confidence': sentiment['score'],
            'keywords': self.extract_key_topics(keywords),
            'fake_score': self.detect_fake_review(review_text)
        }

    def aggregate_product_sentiment(self, product_id):
        """제품 전체 리뷰 감정 집계"""
        reviews = self.get_product_reviews(product_id)
        sentiments = [self.analyze_review(r) for r in reviews]

        return {
            'overall_sentiment': self.calculate_weighted_sentiment(sentiments),
            'positive_ratio': len([s for s in sentiments if s['sentiment'] == 'positive']) / len(sentiments),
            'top_keywords': self.get_top_keywords(sentiments),
            'trust_score': self.calculate_trust_score(sentiments)
        }
```

**데모 가치:**
- NLP/Transformers 실전 활용
- 실시간 스트림 처리 (Kafka)
- 감정 분석 결과를 제품 랭킹에 반영

---

### 3. Dynamic Pricing Service (AI 기반 동적 가격 책정)

**기술 스택:** Python, Scikit-learn, XGBoost, Redis

**기능:**
- 수요/공급 기반 실시간 가격 조정
- 경쟁사 가격 모니터링
- 고객 세그먼트별 맞춤 가격
- A/B 테스트 기반 최적 가격 탐색

**구현 예시:**
```python
import xgboost as xgb
from datetime import datetime

class DynamicPricingService:
    def __init__(self):
        self.pricing_model = self.load_model()
        self.price_cache = Redis()

    def calculate_optimal_price(self, product_id, context):
        """최적 가격 계산"""
        features = self.extract_features(product_id, context)

        # Features:
        # - 현재 재고량
        # - 최근 판매 속도
        # - 시간대 (피크/비피크)
        # - 경쟁사 가격
        # - 고객 세그먼트 (신규/단골, VIP 등)
        # - 계절성
        # - 특별 이벤트 (블랙프라이데이 등)

        base_price = self.get_base_price(product_id)
        optimal_price = self.pricing_model.predict([features])[0]

        # 가격 변동 제한 (너무 급격한 변화 방지)
        min_price = base_price * 0.8
        max_price = base_price * 1.5

        final_price = np.clip(optimal_price, min_price, max_price)

        return {
            'product_id': product_id,
            'base_price': base_price,
            'dynamic_price': final_price,
            'discount_percentage': (base_price - final_price) / base_price * 100,
            'reason': self.explain_pricing(features),
            'expires_at': datetime.now() + timedelta(minutes=15)
        }

    def extract_features(self, product_id, context):
        """가격 책정을 위한 특징 추출"""
        return {
            'inventory_level': self.get_inventory(product_id),
            'sales_velocity': self.get_recent_sales_rate(product_id),
            'time_of_day': datetime.now().hour,
            'day_of_week': datetime.now().weekday(),
            'competitor_price': self.get_competitor_price(product_id),
            'customer_segment': context.get('customer_segment', 'regular'),
            'cart_value': context.get('cart_total', 0),
            'page_views_24h': self.get_product_views(product_id, hours=24)
        }
```

**데모 가치:**
- ML 기반 비즈니스 의사결정
- 실시간 데이터 처리
- 가격 최적화 알고리즘

---

## 🔄 실시간 처리 서비스

### 4. Real-time Inventory Sync Service (실시간 재고 동기화)

**기술 스택:** Go, Kafka, WebSocket, PostgreSQL

**기능:**
- 여러 창고 간 실시간 재고 동기화
- 재고 임계값 도달 시 자동 알림
- 재고 예측 (머신러닝 기반)
- 실시간 재고 현황 대시보드

**구현 예시:**
```go
// inventorysyncservice/inventory_sync.go
package main

import (
    "context"
    "github.com/segmentio/kafka-go"
    "github.com/gorilla/websocket"
)

type InventorySyncService struct {
    kafkaReader *kafka.Reader
    wsClients   map[string]*websocket.Conn
    inventory   map[string]int
}

func (s *InventorySyncService) StreamInventoryUpdates(ctx context.Context) {
    for {
        msg, err := s.kafkaReader.ReadMessage(ctx)
        if err != nil {
            continue
        }

        var update InventoryUpdate
        json.Unmarshal(msg.Value, &update)

        // 재고 업데이트
        s.updateInventory(update)

        // 임계값 체크
        if s.isLowStock(update.ProductID) {
            s.sendAlert(update.ProductID)
        }

        // WebSocket으로 실시간 브로드캐스트
        s.broadcastToClients(update)
    }
}

func (s *InventorySyncService) PredictStockout(productID string) time.Time {
    """재고 소진 시점 예측"""
    salesRate := s.calculateSalesVelocity(productID)
    currentStock := s.inventory[productID]

    daysUntilStockout := float64(currentStock) / salesRate
    return time.Now().Add(time.Duration(daysUntilStockout * 24) * time.Hour)
}

func (s *InventorySyncService) AutoReorder(productID string) {
    """자동 재발주"""
    prediction := s.PredictStockout(productID)

    if prediction.Sub(time.Now()).Hours() < 72 { // 3일 이내 소진 예상
        s.createPurchaseOrder(productID)
    }
}
```

**데모 가치:**
- Event-driven architecture
- WebSocket 실시간 통신
- Kafka 스트림 처리
- 예측 알고리즘

---

### 5. Live Shopping Stream Service (라이브 커머스)

**기술 스택:** Node.js, WebRTC, Redis, MongoDB

**기능:**
- 실시간 비디오 스트리밍
- 라이브 채팅
- 실시간 제품 표시 및 구매
- 타임딜/한정 수량 판매
- 시청자 참여 게임/이벤트

**구현 예시:**
```javascript
// livestreamservice/stream_manager.js
class LiveStreamService {
    constructor() {
        this.activeStreams = new Map();
        this.redis = new Redis();
        this.webrtc = new WebRTCManager();
    }

    async startStream(hostId, streamConfig) {
        const streamId = generateStreamId();

        // WebRTC 연결 설정
        const stream = await this.webrtc.createStream({
            host: hostId,
            quality: streamConfig.quality,
            maxViewers: streamConfig.maxViewers
        });

        // 실시간 채팅방 생성
        const chatRoom = await this.createChatRoom(streamId);

        // 실시간 통계 추적
        this.trackMetrics(streamId);

        this.activeStreams.set(streamId, {
            stream,
            chatRoom,
            viewers: new Set(),
            products: [],
            startTime: new Date()
        });

        return streamId;
    }

    async addProductToStream(streamId, product) {
        """스트림에 제품 추가"""
        const stream = this.activeStreams.get(streamId);

        stream.products.push(product);

        // 모든 시청자에게 알림
        this.broadcast(streamId, {
            type: 'PRODUCT_ADDED',
            product: product,
            specialOffer: this.generateLimitedOffer(product)
        });
    }

    async handleInstantPurchase(streamId, userId, productId) {
        """라이브 중 즉시 구매"""
        const timeLimit = 60; // 60초 한정

        const order = await this.createFlashOrder({
            userId,
            productId,
            streamId,
            expiresIn: timeLimit
        });

        // 실시간 구매 알림
        this.broadcast(streamId, {
            type: 'SOMEONE_PURCHASED',
            product: productId,
            remainingStock: await this.getStock(productId)
        });

        return order;
    }

    trackMetrics(streamId) {
        """실시간 메트릭 추적"""
        setInterval(async () => {
            const metrics = {
                viewers: this.activeStreams.get(streamId).viewers.size,
                engagement: await this.calculateEngagement(streamId),
                sales: await this.getStreamSales(streamId),
                chatActivity: await this.getChatActivity(streamId)
            };

            await this.redis.set(`stream:${streamId}:metrics`, metrics);
            this.broadcast(streamId, { type: 'METRICS_UPDATE', metrics });
        }, 5000);
    }
}
```

**데모 가치:**
- WebRTC 실시간 미디어
- 대규모 동시 접속 처리
- 실시간 이벤트 처리

---

## 🎮 게이미피케이션 서비스

### 6. Loyalty & Gamification Service (충성도 및 게임화)

**기술 스택:** Java/Kotlin, Spring Boot, PostgreSQL, Redis

**기능:**
- 포인트/배지/레벨 시스템
- 일일 미션 및 도전과제
- 친구 초대 보상
- 리더보드
- 가상 아바타/커스터마이징
- 스핀 룰렛, 출석 체크 등

**구현 예시:**
```kotlin
// gamificationservice/GamificationService.kt
@Service
class GamificationService {

    data class UserProgress(
        val userId: String,
        val level: Int,
        val xp: Int,
        val points: Int,
        val badges: List<Badge>,
        val streak: Int
    )

    fun awardPoints(userId: String, action: UserAction): PointsReward {
        val points = calculatePoints(action)
        val multiplier = getMultiplier(userId) // 레벨, 스트릭 등에 따른 배수

        val totalPoints = (points * multiplier).toInt()

        updateUserPoints(userId, totalPoints)
        checkForLevelUp(userId)
        checkForBadges(userId, action)

        return PointsReward(
            points = totalPoints,
            reason = action.description,
            multiplier = multiplier,
            newBadges = checkNewBadges(userId)
        )
    }

    fun getDailyMissions(userId: String): List<Mission> {
        val userProfile = getUserProfile(userId)

        return listOf(
            Mission(
                id = "daily_purchase",
                title = "첫 구매하기",
                description = "오늘 첫 제품을 구매하세요",
                reward = 100,
                progress = userProfile.todayPurchases,
                target = 1
            ),
            Mission(
                id = "review_write",
                title = "리뷰 작성하기",
                description = "구매한 제품에 리뷰를 남기세요",
                reward = 50,
                progress = userProfile.todayReviews,
                target = 1
            ),
            Mission(
                id = "share_product",
                title = "제품 공유하기",
                description = "마음에 드는 제품을 친구에게 공유하세요",
                reward = 30,
                progress = userProfile.todayShares,
                target = 3
            )
        )
    }

    fun spinLuckyWheel(userId: String): WheelReward {
        """행운의 룰렛"""
        if (!canSpin(userId)) {
            throw InsufficientPointsException()
        }

        deductPoints(userId, SPIN_COST)

        val reward = generateWeightedReward(
            rewards = listOf(
                Reward("points", 10, weight = 40),
                Reward("points", 50, weight = 30),
                Reward("points", 100, weight = 20),
                Reward("discount", "10%", weight = 7),
                Reward("discount", "20%", weight = 2),
                Reward("free_shipping", null, weight = 1)
            )
        )

        applyReward(userId, reward)

        return WheelReward(
            type = reward.type,
            value = reward.value,
            description = getRewardDescription(reward)
        )
    }

    fun getLeaderboard(type: LeaderboardType, period: Period): List<LeaderboardEntry> {
        """리더보드"""
        return when(type) {
            LeaderboardType.POINTS -> getTopPointsEarners(period)
            LeaderboardType.PURCHASES -> getTopBuyers(period)
            LeaderboardType.REVIEWS -> getTopReviewers(period)
            LeaderboardType.REFERRALS -> getTopReferrers(period)
        }
    }
}
```

**데모 가치:**
- 사용자 참여 증대 패턴
- 복잡한 비즈니스 로직 처리
- 실시간 진행상황 추적

---

## 🌐 소셜/커뮤니티 서비스

### 7. Social Shopping Service (소셜 쇼핑)

**기술 스택:** Node.js, GraphQL, Neo4j (그래프 DB), Redis

**기능:**
- 친구 팔로우/팔로워
- 위시리스트 공유
- 제품 피드 (인스타그램 스타일)
- 공동 구매 (그룹 바잉)
- 소셜 증명 ("친구 10명이 이 제품을 구매했습니다")

**구현 예시:**
```javascript
// socialshoppingservice/social_service.js
class SocialShoppingService {
    constructor() {
        this.neo4j = new Neo4jDriver();
        this.redis = new Redis();
    }

    async getFriendActivity(userId, limit = 20) {
        """친구들의 최근 활동"""
        const query = `
            MATCH (user:User {id: $userId})-[:FOLLOWS]->(friend:User)
            MATCH (friend)-[r:PURCHASED|REVIEWED|LIKED]->(product:Product)
            RETURN friend, type(r) as action, product, r.timestamp as timestamp
            ORDER BY timestamp DESC
            LIMIT $limit
        `;

        const activities = await this.neo4j.run(query, { userId, limit });

        return activities.map(a => ({
            friend: a.friend,
            action: a.action,
            product: a.product,
            timestamp: a.timestamp,
            socialProof: this.generateSocialProof(a)
        }));
    }

    async createGroupBuy(userId, productId, targetCount, deadline) {
        """공동 구매 생성"""
        const groupBuy = {
            id: generateId(),
            creator: userId,
            product: productId,
            participants: [userId],
            targetCount: targetCount,
            currentCount: 1,
            pricePerPerson: this.calculateGroupPrice(productId, targetCount),
            deadline: deadline,
            status: 'ACTIVE'
        };

        await this.saveGroupBuy(groupBuy);

        // 친구들에게 알림
        await this.notifyFriends(userId, {
            type: 'GROUP_BUY_INVITATION',
            groupBuy: groupBuy
        });

        return groupBuy;
    }

    async joinGroupBuy(userId, groupBuyId) {
        const groupBuy = await this.getGroupBuy(groupBuyId);

        if (groupBuy.currentCount >= groupBuy.targetCount) {
            throw new Error('Group buy is full');
        }

        groupBuy.participants.push(userId);
        groupBuy.currentCount++;

        if (groupBuy.currentCount === groupBuy.targetCount) {
            await this.completeGroupBuy(groupBuyId);
        }

        return groupBuy;
    }

    async getProductSocialProof(productId, userId) {
        """소셜 증명 데이터"""
        const query = `
            MATCH (user:User {id: $userId})-[:FOLLOWS]->(friend:User)
            MATCH (friend)-[:PURCHASED]->(product:Product {id: $productId})
            RETURN friend
        `;

        const friends = await this.neo4j.run(query, { userId, productId });

        return {
            friendsPurchased: friends.length,
            friendNames: friends.slice(0, 3).map(f => f.name),
            totalPurchases: await this.getTotalPurchases(productId),
            averageRating: await this.getAverageRating(productId),
            trendingScore: await this.calculateTrendingScore(productId)
        };
    }

    async createProductFeed(userId) {
        """개인화된 제품 피드"""
        const [
            friendsLiked,
            trending,
            recommendations,
            newArrivals
        ] = await Promise.all([
            this.getFriendsLikedProducts(userId),
            this.getTrendingProducts(),
            this.getPersonalizedRecommendations(userId),
            this.getNewArrivals()
        ]);

        // 피드 알고리즘으로 믹스
        const feed = this.mixFeedItems([
            ...friendsLiked.map(p => ({ ...p, source: 'friends' })),
            ...trending.map(p => ({ ...p, source: 'trending' })),
            ...recommendations.map(p => ({ ...p, source: 'for_you' })),
            ...newArrivals.map(p => ({ ...p, source: 'new' }))
        ]);

        return feed;
    }
}
```

**데모 가치:**
- 그래프 데이터베이스 활용
- 소셜 네트워크 패턴
- 개인화 알고리즘

---

## 🔍 분석 및 인사이트 서비스

### 8. Customer Analytics Service (고객 분석)

**기술 스택:** Python, Apache Spark, ClickHouse, Superset

**기능:**
- 실시간 사용자 행동 추적
- 코호트 분석
- 퍼널 분석
- RFM (Recency, Frequency, Monetary) 분석
- 이탈 예측
- 생애 가치(LTV) 예측

**구현 예시:**
```python
# analyticsservice/customer_analytics.py
from pyspark.sql import SparkSession
import pandas as pd
from sklearn.ensemble import RandomForestClassifier

class CustomerAnalyticsService:
    def __init__(self):
        self.spark = SparkSession.builder.appName("CustomerAnalytics").getOrCreate()
        self.churn_model = self.load_churn_model()

    def analyze_user_journey(self, user_id):
        """사용자 여정 분석"""
        events = self.get_user_events(user_id)

        journey = {
            'acquisition_channel': self.get_acquisition_channel(user_id),
            'first_purchase_time': self.time_to_first_purchase(user_id),
            'conversion_funnel': self.analyze_funnel(events),
            'touchpoints': self.identify_touchpoints(events),
            'drop_off_points': self.find_drop_offs(events)
        }

        return journey

    def segment_customers(self, method='rfm'):
        """고객 세그먼테이션"""
        if method == 'rfm':
            return self.rfm_segmentation()
        elif method == 'behavioral':
            return self.behavioral_clustering()
        elif method == 'value':
            return self.value_based_segmentation()

    def rfm_segmentation(self):
        """RFM 분석"""
        query = """
            SELECT
                user_id,
                DATEDIFF(CURRENT_DATE, MAX(order_date)) as recency,
                COUNT(DISTINCT order_id) as frequency,
                SUM(order_total) as monetary
            FROM orders
            GROUP BY user_id
        """

        rfm_df = self.spark.sql(query).toPandas()

        # RFM 점수 계산 (1-5)
        rfm_df['R_Score'] = pd.qcut(rfm_df['recency'], 5, labels=[5,4,3,2,1])
        rfm_df['F_Score'] = pd.qcut(rfm_df['frequency'], 5, labels=[1,2,3,4,5])
        rfm_df['M_Score'] = pd.qcut(rfm_df['monetary'], 5, labels=[1,2,3,4,5])

        # 세그먼트 라벨링
        def assign_segment(row):
            if row['R_Score'] >= 4 and row['F_Score'] >= 4:
                return 'Champions'
            elif row['R_Score'] >= 3 and row['F_Score'] >= 3:
                return 'Loyal Customers'
            elif row['R_Score'] >= 4 and row['F_Score'] <= 2:
                return 'Promising'
            elif row['R_Score'] <= 2 and row['F_Score'] >= 4:
                return 'At Risk'
            elif row['R_Score'] <= 2 and row['F_Score'] <= 2:
                return 'Lost'
            else:
                return 'Need Attention'

        rfm_df['Segment'] = rfm_df.apply(assign_segment, axis=1)

        return rfm_df

    def predict_churn(self, user_id):
        """이탈 예측"""
        features = self.extract_churn_features(user_id)

        churn_probability = self.churn_model.predict_proba([features])[0][1]

        return {
            'user_id': user_id,
            'churn_probability': churn_probability,
            'risk_level': 'HIGH' if churn_probability > 0.7 else 'MEDIUM' if churn_probability > 0.4 else 'LOW',
            'recommended_actions': self.get_retention_actions(churn_probability),
            'key_factors': self.explain_churn_factors(features)
        }

    def calculate_ltv(self, user_id):
        """고객 생애 가치 예측"""
        historical_orders = self.get_user_orders(user_id)

        # 구매 빈도 예측
        avg_days_between_purchases = self.calculate_purchase_interval(historical_orders)
        estimated_purchases_per_year = 365 / avg_days_between_purchases

        # 평균 주문 금액
        avg_order_value = sum(o['total'] for o in historical_orders) / len(historical_orders)

        # 고객 수명 예측 (년)
        estimated_lifetime = self.predict_customer_lifetime(user_id)

        ltv = estimated_purchases_per_year * avg_order_value * estimated_lifetime

        return {
            'ltv': ltv,
            'avg_order_value': avg_order_value,
            'purchase_frequency': estimated_purchases_per_year,
            'estimated_lifetime_years': estimated_lifetime,
            'confidence': self.calculate_confidence(historical_orders)
        }

    def create_cohort_analysis(self, cohort_type='monthly'):
        """코호트 분석"""
        query = """
            WITH user_cohorts AS (
                SELECT
                    user_id,
                    DATE_TRUNC('month', MIN(order_date)) as cohort_month
                FROM orders
                GROUP BY user_id
            )
            SELECT
                cohort_month,
                DATE_TRUNC('month', order_date) as activity_month,
                COUNT(DISTINCT o.user_id) as active_users,
                SUM(order_total) as revenue
            FROM orders o
            JOIN user_cohorts uc ON o.user_id = uc.user_id
            GROUP BY cohort_month, activity_month
        """

        cohorts = self.spark.sql(query).toPandas()

        # 리텐션 매트릭스 생성
        retention_matrix = self.build_retention_matrix(cohorts)

        return {
            'retention_matrix': retention_matrix,
            'visualization': self.create_cohort_heatmap(retention_matrix)
        }
```

**데모 가치:**
- Big Data 처리 (Spark)
- ML 기반 예측
- 비즈니스 인텔리전스

---

## 🔐 블록체인/Web3 서비스

### 9. NFT Collectibles Service (NFT 수집품)

**기술 스택:** Solidity, Hardhat, Ethers.js, IPFS, Node.js

**기능:**
- 한정판 제품 NFT 발행
- NFT 소유자 전용 혜택
- NFT 마켓플레이스
- 디지털 인증서
- 메타버스 연동

**구현 예시:**
```solidity
// nftservice/contracts/ProductNFT.sol
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

contract ProductNFT is ERC721 {
    using Counters for Counters.Counter;
    Counters.Counter private _tokenIds;

    struct ProductMetadata {
        string productId;
        string name;
        uint256 edition;
        uint256 totalEditions;
        string ipfsHash;
        uint256 purchaseTimestamp;
    }

    mapping(uint256 => ProductMetadata) public nftMetadata;
    mapping(string => uint256[]) public productToTokens;

    event NFTMinted(
        uint256 indexed tokenId,
        address indexed owner,
        string productId,
        uint256 edition
    );

    constructor() ERC721("LimitedEditionProduct", "LEP") {}

    function mintProductNFT(
        address buyer,
        string memory productId,
        string memory name,
        uint256 edition,
        uint256 totalEditions,
        string memory ipfsHash
    ) public returns (uint256) {
        _tokenIds.increment();
        uint256 newTokenId = _tokenIds.current();

        _safeMint(buyer, newTokenId);

        nftMetadata[newTokenId] = ProductMetadata({
            productId: productId,
            name: name,
            edition: edition,
            totalEditions: totalEditions,
            ipfsHash: ipfsHash,
            purchaseTimestamp: block.timestamp
        });

        productToTokens[productId].push(newTokenId);

        emit NFTMinted(newTokenId, buyer, productId, edition);

        return newTokenId;
    }

    function getOwnerBenefits(uint256 tokenId) public view returns (string[] memory) {
        require(_exists(tokenId), "Token does not exist");

        ProductMetadata memory metadata = nftMetadata[tokenId];
        string[] memory benefits = new string[](3);

        benefits[0] = "Lifetime 10% discount on all purchases";
        benefits[1] = "Early access to new product launches";
        benefits[2] = "Exclusive NFT holder events";

        // 에디션 번호가 작을수록 더 많은 혜택
        if (metadata.edition <= 10) {
            // benefits.push("VIP customer service");
        }

        return benefits;
    }

    function verifyAuthenticity(uint256 tokenId) public view returns (bool) {
        return _exists(tokenId);
    }
}
```

```javascript
// nftservice/nft_service.js
const { ethers } = require('ethers');
const { create } = require('ipfs-http-client');

class NFTService {
    constructor() {
        this.provider = new ethers.providers.JsonRpcProvider(process.env.RPC_URL);
        this.contract = new ethers.Contract(CONTRACT_ADDRESS, ABI, this.provider);
        this.ipfs = create({ host: 'ipfs.infura.io', port: 5001, protocol: 'https' });
    }

    async createLimitedEditionNFT(productId, buyer, edition, totalEditions) {
        """한정판 제품 NFT 발행"""

        // 메타데이터를 IPFS에 업로드
        const metadata = {
            name: `Limited Edition ${productId} #${edition}`,
            description: `This is ${edition} of ${totalEditions} limited edition`,
            image: await this.uploadImageToIPFS(productId),
            attributes: [
                { trait_type: "Edition", value: edition },
                { trait_type: "Total Supply", value: totalEditions },
                { trait_type: "Product ID", value: productId },
                { trait_type: "Rarity", value: this.calculateRarity(edition, totalEditions) }
            ]
        };

        const ipfsHash = await this.uploadMetadataToIPFS(metadata);

        // NFT 발행
        const tx = await this.contract.mintProductNFT(
            buyer,
            productId,
            metadata.name,
            edition,
            totalEditions,
            ipfsHash
        );

        await tx.wait();

        return {
            tokenId: await this.getLatestTokenId(),
            ipfsUrl: `https://ipfs.io/ipfs/${ipfsHash}`,
            openseaUrl: this.getOpenSeaUrl(tokenId)
        };
    }

    async getOwnerPerks(walletAddress) {
        """NFT 소유자 혜택 조회"""
        const tokens = await this.contract.tokensOfOwner(walletAddress);

        const perks = {
            discountRate: 0,
            earlyAccess: false,
            exclusiveEvents: false,
            totalNFTs: tokens.length
        };

        if (tokens.length > 0) {
            perks.discountRate = 10; // 기본 10%
            perks.earlyAccess = true;
        }

        if (tokens.length >= 5) {
            perks.discountRate = 20; // 5개 이상 소유 시 20%
            perks.exclusiveEvents = true;
        }

        return perks;
    }
}
```

**데모 가치:**
- 블록체인 통합
- 스마트 컨트랙트
- Web3 인증
- IPFS 분산 스토리지

---

## 🎯 IoT/엣지 서비스

### 10. Smart Mirror Service (스마트 미러 가상 피팅)

**기술 스택:** Python, OpenCV, MediaPipe, TensorFlow, WebSocket

**기능:**
- 실시간 AR 가상 피팅
- 얼굴/신체 인식
- 옷/액세서리 오버레이
- 사이즈 추천
- 스타일 조합 제안

**구현 예시:**
```python
# smartmirrorservice/virtual_fitting.py
import cv2
import mediapipe as mp
import numpy as np

class VirtualFittingService:
    def __init__(self):
        self.mp_pose = mp.solutions.pose
        self.mp_face = mp.solutions.face_mesh
        self.pose = self.mp_pose.Pose()
        self.face_mesh = self.mp_face.FaceMesh()

    def process_frame(self, frame, product_type, product_image):
        """프레임 처리 및 가상 피팅"""
        results_pose = self.pose.process(cv2.cvtColor(frame, cv2.COLOR_BGR2RGB))
        results_face = self.face_mesh.process(cv2.cvtColor(frame, cv2.COLOR_BGR2RGB))

        if product_type == 'clothing':
            return self.overlay_clothing(frame, results_pose, product_image)
        elif product_type == 'glasses':
            return self.overlay_glasses(frame, results_face, product_image)
        elif product_type == 'hat':
            return self.overlay_hat(frame, results_face, product_image)

        return frame

    def overlay_glasses(self, frame, face_results, glasses_image):
        """안경 오버레이"""
        if not face_results.multi_face_landmarks:
            return frame

        landmarks = face_results.multi_face_landmarks[0]

        # 얼굴 랜드마크에서 눈 위치 추출
        left_eye = landmarks.landmark[33]  # 왼쪽 눈
        right_eye = landmarks.landmark[263]  # 오른쪽 눈

        # 안경 크기 및 각도 계산
        eye_distance = np.linalg.norm([
            left_eye.x - right_eye.x,
            left_eye.y - right_eye.y
        ])

        # 안경 이미지 변환
        glasses_width = int(eye_distance * frame.shape[1] * 2.5)
        glasses_resized = cv2.resize(glasses_image, (glasses_width, glasses_width // 3))

        # 각도 조정
        angle = np.degrees(np.arctan2(right_eye.y - left_eye.y, right_eye.x - left_eye.x))
        glasses_rotated = self.rotate_image(glasses_resized, angle)

        # 오버레이
        overlay_frame = self.overlay_image(frame, glasses_rotated, left_eye, right_eye)

        return overlay_frame

    def measure_body_dimensions(self, pose_results, user_height_cm):
        """신체 치수 측정"""
        if not pose_results.pose_landmarks:
            return None

        landmarks = pose_results.pose_landmarks.landmark

        # 어깨 너비
        left_shoulder = landmarks[self.mp_pose.PoseLandmark.LEFT_SHOULDER]
        right_shoulder = landmarks[self.mp_pose.PoseLandmark.RIGHT_SHOULDER]
        shoulder_width_px = np.linalg.norm([
            left_shoulder.x - right_shoulder.x,
            left_shoulder.y - right_shoulder.y
        ])

        # 픽셀을 실제 cm로 변환 (사용자가 입력한 키 기준)
        pixel_to_cm = user_height_cm / self.calculate_body_height_px(landmarks)

        measurements = {
            'shoulder_width': shoulder_width_px * pixel_to_cm,
            'chest_width': self.estimate_chest_width(landmarks) * pixel_to_cm,
            'waist_width': self.estimate_waist_width(landmarks) * pixel_to_cm,
            'hip_width': self.estimate_hip_width(landmarks) * pixel_to_cm,
        }

        return measurements

    def recommend_size(self, measurements, product_id):
        """사이즈 추천"""
        product_sizing = self.get_product_sizing(product_id)

        # 각 사이즈와의 차이 계산
        size_scores = {}
        for size, size_measurements in product_sizing.items():
            score = 0
            for key in measurements:
                if key in size_measurements:
                    diff = abs(measurements[key] - size_measurements[key])
                    score += diff
            size_scores[size] = score

        # 가장 적합한 사이즈
        recommended_size = min(size_scores, key=size_scores.get)

        return {
            'recommended_size': recommended_size,
            'fit_score': 100 - (size_scores[recommended_size] / 10),  # 0-100 점수
            'alternative_sizes': self.get_alternative_sizes(size_scores, recommended_size)
        }
```

**데모 가치:**
- Computer Vision 실전 활용
- AR 기술
- 실시간 처리
- IoT 디바이스 통합

---

## 🤝 협업 서비스

### 11. Collaborative Shopping Service (협업 쇼핑)

**기술 스택:** Node.js, Socket.io, WebRTC, React

**기능:**
- 화면 공유 쇼핑
- 음성/비디오 채팅
- 공동 장바구니
- 실시간 투표
- 함께 결제하기

**구현 예시:**
```javascript
// collaborativeshoppingservice/collaborative_session.js
class CollaborativeShoppingSession {
    constructor(sessionId) {
        this.sessionId = sessionId;
        this.participants = new Map();
        this.sharedCart = [];
        this.currentProduct = null;
        this.votes = new Map();
    }

    addParticipant(userId, socket) {
        """참가자 추가"""
        this.participants.set(userId, {
            socket: socket,
            cursor: { x: 0, y: 0 },
            status: 'active',
            role: this.participants.size === 0 ? 'host' : 'guest'
        });

        // 모든 참가자에게 알림
        this.broadcast('PARTICIPANT_JOINED', {
            userId: userId,
            totalParticipants: this.participants.size
        });

        // 새 참가자에게 현재 상태 동기화
        socket.emit('SESSION_STATE', {
            cart: this.sharedCart,
            currentProduct: this.currentProduct,
            participants: Array.from(this.participants.keys())
        });
    }

    syncCursor(userId, position) {
        """커서 위치 동기화"""
        const participant = this.participants.get(userId);
        if (participant) {
            participant.cursor = position;

            // 다른 참가자들에게 브로드캐스트
            this.broadcast('CURSOR_MOVE', {
                userId: userId,
                position: position
            }, userId);
        }
    }

    addToSharedCart(userId, product) {
        """공동 장바구니에 추가"""
        this.sharedCart.push({
            ...product,
            addedBy: userId,
            timestamp: new Date()
        });

        this.broadcast('CART_UPDATED', {
            cart: this.sharedCart,
            addedBy: userId,
            product: product
        });
    }

    startProductVote(userId, product) {
        """제품 투표 시작"""
        this.votes.clear();
        this.currentProduct = product;

        this.broadcast('VOTE_STARTED', {
            product: product,
            initiatedBy: userId,
            deadline: Date.now() + 60000 // 1분
        });

        // 1분 후 자동 집계
        setTimeout(() => {
            this.tallyVotes();
        }, 60000);
    }

    castVote(userId, vote) {
        """투표하기"""
        this.votes.set(userId, vote); // 'yes', 'no', 'maybe'

        this.broadcast('VOTE_CAST', {
            userId: userId,
            totalVotes: this.votes.size,
            requiredVotes: this.participants.size
        });

        // 모두 투표했으면 즉시 집계
        if (this.votes.size === this.participants.size) {
            this.tallyVotes();
        }
    }

    tallyVotes() {
        """투표 집계"""
        const results = {
            yes: 0,
            no: 0,
            maybe: 0
        };

        this.votes.forEach(vote => {
            results[vote]++;
        });

        const decision = results.yes > results.no ? 'approved' : 'rejected';

        this.broadcast('VOTE_COMPLETE', {
            results: results,
            decision: decision,
            product: this.currentProduct
        });

        if (decision === 'approved') {
            this.addToSharedCart('group', this.currentProduct);
        }

        this.votes.clear();
        this.currentProduct = null;
    }

    splitPayment() {
        """결제 분할"""
        const total = this.sharedCart.reduce((sum, item) => sum + item.price, 0);
        const perPerson = total / this.participants.size;

        const splits = [];
        this.participants.forEach((participant, userId) => {
            splits.push({
                userId: userId,
                amount: perPerson,
                items: this.sharedCart.map(item => ({
                    ...item,
                    splitPrice: item.price / this.participants.size
                }))
            });
        });

        this.broadcast('PAYMENT_SPLIT', {
            total: total,
            perPerson: perPerson,
            splits: splits
        });

        return splits;
    }

    broadcast(event, data, excludeUserId = null) {
        """모든 참가자에게 브로드캐스트"""
        this.participants.forEach((participant, userId) => {
            if (userId !== excludeUserId) {
                participant.socket.emit(event, data);
            }
        });
    }
}

// WebSocket 서버
io.on('connection', (socket) => {
    socket.on('JOIN_SESSION', ({ sessionId, userId }) => {
        const session = getOrCreateSession(sessionId);
        session.addParticipant(userId, socket);
    });

    socket.on('CURSOR_MOVE', ({ sessionId, userId, position }) => {
        const session = sessions.get(sessionId);
        session?.syncCursor(userId, position);
    });

    socket.on('ADD_TO_CART', ({ sessionId, userId, product }) => {
        const session = sessions.get(sessionId);
        session?.addToSharedCart(userId, product);
    });

    socket.on('START_VOTE', ({ sessionId, userId, product }) => {
        const session = sessions.get(sessionId);
        session?.startProductVote(userId, product);
    });

    socket.on('CAST_VOTE', ({ sessionId, userId, vote }) => {
        const session = sessions.get(sessionId);
        session?.castVote(userId, vote);
    });
});
```

**데모 가치:**
- 실시간 협업 패턴
- WebSocket/WebRTC
- 동시성 제어

---

## 📱 모바일 우선 서비스

### 12. Progressive Web App Service (PWA 최적화)

**기술 스택:** Service Workers, IndexedDB, Web Push API

**기능:**
- 오프라인 모드
- 백그라운드 동기화
- 푸시 알림
- 앱 설치 프롬프트
- 데이터 절약 모드

**구현 예시:**
```javascript
// pwaservice/service-worker.js
const CACHE_VERSION = 'v1';
const CACHE_STATIC = `static-${CACHE_VERSION}`;
const CACHE_DYNAMIC = `dynamic-${CACHE_VERSION}`;
const CACHE_PRODUCTS = `products-${CACHE_VERSION}`;

// 설치 시 정적 자산 캐싱
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_STATIC).then((cache) => {
            return cache.addAll([
                '/',
                '/css/main.css',
                '/js/app.js',
                '/images/logo.png',
                '/offline.html'
            ]);
        })
    );
});

// 네트워크 요청 가로채기
self.addEventListener('fetch', (event) => {
    const { request } = event;

    // API 요청 처리
    if (request.url.includes('/api/')) {
        event.respondWith(networkFirstStrategy(request));
    }
    // 제품 이미지 처리
    else if (request.url.includes('/products/')) {
        event.respondWith(cacheFirstStrategy(request));
    }
    // 정적 자산 처리
    else {
        event.respondWith(staleWhileRevalidateStrategy(request));
    }
});

// Network First 전략 (API용)
async function networkFirstStrategy(request) {
    try {
        const networkResponse = await fetch(request);
        const cache = await caches.open(CACHE_DYNAMIC);
        cache.put(request, networkResponse.clone());
        return networkResponse;
    } catch (error) {
        const cachedResponse = await caches.match(request);
        return cachedResponse || caches.match('/offline.html');
    }
}

// Cache First 전략 (이미지용)
async function cacheFirstStrategy(request) {
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
        return cachedResponse;
    }

    try {
        const networkResponse = await fetch(request);
        const cache = await caches.open(CACHE_PRODUCTS);
        cache.put(request, networkResponse.clone());
        return networkResponse;
    } catch (error) {
        return caches.match('/images/placeholder.png');
    }
}

// Stale While Revalidate 전략
async function staleWhileRevalidateStrategy(request) {
    const cachedResponse = await caches.match(request);

    const fetchPromise = fetch(request).then((networkResponse) => {
        const cache = caches.open(CACHE_STATIC);
        cache.then((c) => c.put(request, networkResponse.clone()));
        return networkResponse;
    });

    return cachedResponse || fetchPromise;
}

// 백그라운드 동기화
self.addEventListener('sync', (event) => {
    if (event.tag === 'sync-cart') {
        event.waitUntil(syncCart());
    } else if (event.tag === 'sync-orders') {
        event.waitUntil(syncOrders());
    }
});

async function syncCart() {
    """오프라인 중 추가한 장바구니 동기화"""
    const db = await openIndexedDB();
    const pendingItems = await db.getAll('pendingCartItems');

    for (const item of pendingItems) {
        try {
            await fetch('/api/cart', {
                method: 'POST',
                body: JSON.stringify(item),
                headers: { 'Content-Type': 'application/json' }
            });
            await db.delete('pendingCartItems', item.id);
        } catch (error) {
            console.error('Failed to sync cart item:', error);
        }
    }
}

// 푸시 알림
self.addEventListener('push', (event) => {
    const data = event.data.json();

    const options = {
        body: data.body,
        icon: '/images/icon.png',
        badge: '/images/badge.png',
        vibrate: [200, 100, 200],
        data: {
            url: data.url
        },
        actions: [
            { action: 'view', title: '보기' },
            { action: 'dismiss', title: '닫기' }
        ]
    };

    event.waitUntil(
        self.registration.showNotification(data.title, options)
    );
});

// 알림 클릭 처리
self.addEventListener('notificationclick', (event) => {
    event.notification.close();

    if (event.action === 'view') {
        event.waitUntil(
            clients.openWindow(event.notification.data.url)
        );
    }
});
```

```javascript
// pwaservice/offline_manager.js
class OfflineManager {
    constructor() {
        this.db = null;
        this.initDB();
    }

    async initDB() {
        """IndexedDB 초기화"""
        this.db = await openDB('OfflineStore', 1, {
            upgrade(db) {
                db.createObjectStore('pendingCartItems', { keyPath: 'id', autoIncrement: true });
                db.createObjectStore('cachedProducts', { keyPath: 'id' });
                db.createObjectStore('userPreferences', { keyPath: 'key' });
            }
        });
    }

    async addToOfflineCart(product) {
        """오프라인 장바구니에 추가"""
        await this.db.add('pendingCartItems', {
            ...product,
            addedAt: new Date(),
            synced: false
        });

        // 온라인이 되면 동기화
        if ('sync' in self.registration) {
            await self.registration.sync.register('sync-cart');
        }
    }

    async getOfflineCart() {
        """오프라인 장바구니 조회"""
        return await this.db.getAll('pendingCartItems');
    }

    async cacheProduct(product) {
        """제품 정보 캐싱"""
        await this.db.put('cachedProducts', product);
    }

    async getCachedProduct(productId) {
        """캐시된 제품 조회"""
        return await this.db.get('cachedProducts', productId);
    }
}
```

**데모 가치:**
- PWA 베스트 프랙티스
- 오프라인 우선 전략
- 모바일 최적화

---

## 🎨 개인화 서비스

### 13. Hyper-Personalization Service (초개인화)

**기술 스택:** Python, TensorFlow, Redis, Kafka

**기능:**
- 실시간 개인화 홈페이지
- 동적 컨텐츠 생성
- 개인별 가격/프로모션
- 맞춤형 이메일
- 예측 검색

**구현 예시:**
```python
# personalizationservice/hyper_personalization.py
class HyperPersonalizationEngine:
    def __init__(self):
        self.user_profile_model = self.load_model('user_profiling')
        self.content_ranker = self.load_model('content_ranking')
        self.redis = Redis()

    def generate_personalized_homepage(self, user_id, context):
        """초개인화 홈페이지 생성"""
        # 사용자 프로필 로드
        user_profile = self.get_user_profile(user_id)

        # 실시간 컨텍스트
        current_context = {
            'time_of_day': context.get('hour'),
            'device': context.get('device'),
            'location': context.get('location'),
            'weather': self.get_weather(context.get('location')),
            'recent_behavior': self.get_recent_behavior(user_id, hours=24)
        }

        # 홈페이지 섹션 생성
        sections = []

        # 1. 히어로 배너 (최고 우선순위)
        hero_banner = self.select_hero_banner(user_profile, current_context)
        sections.append(hero_banner)

        # 2. 맞춤 추천 ("당신을 위한 추천")
        recommendations = self.get_personalized_recommendations(
            user_id,
            context=current_context,
            count=12
        )
        sections.append({
            'type': 'product_grid',
            'title': self.generate_title(user_profile),  # "오늘의 추천" vs "취향저격 아이템"
            'products': recommendations
        })

        # 3. 시간 기반 컨텐츠
        if current_context['time_of_day'] < 12:
            sections.append(self.get_morning_deals())
        elif current_context['time_of_day'] < 18:
            sections.append(self.get_lunch_specials())
        else:
            sections.append(self.get_evening_picks())

        # 4. 날씨 기반 제안
        if current_context['weather']['temp'] < 10:
            sections.append(self.get_cold_weather_products())
        elif current_context['weather']['condition'] == 'rainy':
            sections.append(self.get_rainy_day_products())

        # 5. 재방문 유도
        abandoned_items = self.get_abandoned_cart(user_id)
        if abandoned_items:
            sections.append({
                'type': 'reminder',
                'title': '장바구니에 남겨두신 상품이 있어요',
                'products': abandoned_items,
                'special_offer': self.generate_incentive(user_profile)  # 쿠폰 등
            })

        # 6. 소셜 증명
        trending_among_similar = self.get_trending_in_segment(user_profile['segment'])
        sections.append({
            'type': 'social_proof',
            'title': '비슷한 취향을 가진 분들이 구매한 상품',
            'products': trending_among_similar
        })

        return {
            'user_id': user_id,
            'generated_at': datetime.now(),
            'sections': sections,
            'ab_tests': self.get_active_ab_tests(user_id)
        }

    def predictive_search(self, user_id, partial_query):
        """예측 검색"""
        user_history = self.get_search_history(user_id)
        popular_searches = self.get_popular_searches()

        # 사용자 히스토리 기반 예측
        user_predictions = self.predict_from_history(partial_query, user_history)

        # 컨텍스트 기반 예측
        context_predictions = self.predict_from_context(partial_query, {
            'season': self.get_current_season(),
            'trending': popular_searches,
            'user_preferences': self.get_user_preferences(user_id)
        })

        # 랭킹 및 병합
        predictions = self.merge_and_rank(user_predictions, context_predictions)

        return predictions[:10]  # Top 10

    def generate_dynamic_email(self, user_id, email_type):
        """동적 이메일 생성"""
        user_profile = self.get_user_profile(user_id)

        # 이메일 발송 최적 시간 예측
        optimal_send_time = self.predict_optimal_send_time(user_id)

        # 컨텐츠 생성
        if email_type == 'promotional':
            products = self.select_products_for_email(user_id, count=6)
            subject = self.generate_subject_line(user_profile, products)

            content = {
                'subject': subject,
                'header_image': self.select_header_image(user_profile),
                'greeting': self.personalize_greeting(user_profile),
                'products': products,
                'cta_text': self.optimize_cta(user_profile),
                'discount': self.calculate_personal_discount(user_id)
            }

        elif email_type == 'cart_abandonment':
            cart_items = self.get_abandoned_cart(user_id)

            content = {
                'subject': f"{user_profile['first_name']}님, 장바구니에 {len(cart_items)}개 상품이 기다리고 있어요",
                'cart_items': cart_items,
                'incentive': self.generate_recovery_incentive(user_id),  # 10% 할인 등
                'urgency': self.create_urgency(cart_items)  # "2개 상품 재고 얼마 남지 않음"
            }

        return {
            'content': content,
            'optimal_send_time': optimal_send_time,
            'ab_variant': self.assign_email_variant(user_id)
        }
```

**데모 가치:**
- 실시간 개인화
- ML 기반 컨텐츠 생성
- 동적 사용자 경험

---

## 📊 우선순위 매트릭스

| 서비스 | 구현 난이도 | 데모 임팩트 | 학습 가치 | 비즈니스 가치 | 추천도 |
|--------|------------|-----------|---------|------------|--------|
| Visual Search | 중 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 🔥🔥🔥🔥🔥 |
| Dynamic Pricing | 중 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 🔥🔥🔥🔥🔥 |
| Real-time Inventory | 낮 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| Sentiment Analysis | 중 | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | 🔥🔥🔥🔥 |
| Live Shopping | 높음 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 🔥🔥🔥 |
| Gamification | 중 | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| Social Shopping | 중 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | 🔥🔥🔥 |
| Customer Analytics | 중 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| NFT Collectibles | 높음 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | 🔥🔥🔥 |
| Smart Mirror | 높음 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | 🔥🔥🔥 |
| Collaborative Shopping | 중 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | 🔥🔥🔥 |
| PWA Service | 낮 | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | 🔥🔥🔥🔥 |
| Hyper-Personalization | 높음 | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 🔥🔥🔥🔥 |

---

## 🚀 빠른 시작 가이드

각 서비스별로 30분 안에 기본 버전을 구현할 수 있는 가이드를 제공할 수 있습니다.

어떤 서비스를 실제로 구현해보시겠습니까?

1. **Visual Search** - 이미지 기반 제품 검색
2. **Dynamic Pricing** - AI 가격 최적화
3. **Gamification** - 포인트/배지 시스템
4. **Real-time Inventory** - 실시간 재고 동기화
5. **PWA Service** - 오프라인 모드

선택하시면 바로 구현 가능한 코드와 함께 프로젝트에 통합하는 방법을 보여드리겠습니다!
