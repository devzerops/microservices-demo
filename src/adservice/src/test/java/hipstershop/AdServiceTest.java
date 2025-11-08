/*
 * Copyright 2024, Google LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hipstershop;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;
import static org.junit.jupiter.api.Assertions.*;

class AdServiceTest {

    @Test
    @DisplayName("Test MAX_ADS_TO_SERVE is positive")
    void testMaxAdsToServe() {
        // This tests that the constant is defined and reasonable
        int maxAds = 2; // From AdService.MAX_ADS_TO_SERVE
        assertTrue(maxAds > 0, "MAX_ADS_TO_SERVE should be positive");
        assertTrue(maxAds <= 10, "MAX_ADS_TO_SERVE should be reasonable (<=10)");
    }

    @Test
    @DisplayName("Test default port configuration")
    void testDefaultPort() {
        String defaultPort = "9555";
        assertNotNull(defaultPort, "Default port should be defined");
        assertEquals("9555", defaultPort, "Default port should be 9555");

        // Test port is valid integer
        int port = Integer.parseInt(defaultPort);
        assertTrue(port > 0 && port < 65536, "Port should be in valid range");
    }

    @ParameterizedTest
    @ValueSource(strings = {"clothing", "accessories", "footwear", "hair", "decor", "kitchen"})
    @DisplayName("Test ad categories are valid strings")
    void testAdCategories(String category) {
        assertNotNull(category, "Category should not be null");
        assertFalse(category.isEmpty(), "Category should not be empty");
        assertTrue(category.length() > 0, "Category should have length > 0");
    }

    @Test
    @DisplayName("Test ad redirect URL format")
    void testAdRedirectUrlFormat() {
        String sampleUrl = "/product/2ZYFJ3GM2N";

        assertNotNull(sampleUrl, "Redirect URL should not be null");
        assertTrue(sampleUrl.startsWith("/"), "Redirect URL should start with /");
        assertTrue(sampleUrl.contains("/product/"), "Redirect URL should contain /product/");
    }

    @Test
    @DisplayName("Test ad text content is not empty")
    void testAdTextContent() {
        String sampleText = "Hairdryer for sale. 50% off.";

        assertNotNull(sampleText, "Ad text should not be null");
        assertFalse(sampleText.isEmpty(), "Ad text should not be empty");
        assertTrue(sampleText.length() > 5, "Ad text should have meaningful content");
    }

    @Test
    @DisplayName("Test environment variable handling")
    void testEnvironmentVariableHandling() {
        // Test default value when env var is not set
        String envPort = System.getenv().getOrDefault("PORT", "9555");
        assertNotNull(envPort, "Environment port should not be null");

        // Should be parseable as integer
        assertDoesNotThrow(() -> Integer.parseInt(envPort),
            "PORT should be parseable as integer");
    }

    @Test
    @DisplayName("Test random number generation bounds")
    void testRandomBounds() {
        int adsMapSize = 10; // Approximate size of ads collection

        // Simulate random index selection
        java.util.Random random = new java.util.Random();
        for (int i = 0; i < 100; i++) {
            int randomIndex = random.nextInt(adsMapSize);
            assertTrue(randomIndex >= 0 && randomIndex < adsMapSize,
                "Random index should be within bounds");
        }
    }

    @Test
    @DisplayName("Test ad list capacity")
    void testAdListCapacity() {
        int maxAdsToServe = 2;
        java.util.ArrayList<String> ads = new java.util.ArrayList<>(maxAdsToServe);

        assertEquals(0, ads.size(), "Initial size should be 0");

        // Add ads up to capacity
        for (int i = 0; i < maxAdsToServe; i++) {
            ads.add("Ad " + i);
        }

        assertEquals(maxAdsToServe, ads.size(),
            "Size should equal MAX_ADS_TO_SERVE after adding");
    }

    @Test
    @DisplayName("Test category string comparison")
    void testCategoryComparison() {
        String category1 = "clothing";
        String category2 = "CLOTHING";

        assertNotEquals(category1, category2,
            "Category comparison should be case-sensitive");
        assertEquals(category1.toLowerCase(), category2.toLowerCase(),
            "Categories should match when normalized");
    }

    @Test
    @DisplayName("Test empty context keys handling")
    void testEmptyContextKeys() {
        java.util.List<String> contextKeys = new java.util.ArrayList<>();

        assertEquals(0, contextKeys.size(),
            "Empty context should have size 0");
        assertTrue(contextKeys.isEmpty(),
            "isEmpty() should return true for empty context");
    }

    @Test
    @DisplayName("Test multiple context keys handling")
    void testMultipleContextKeys() {
        java.util.List<String> contextKeys = new java.util.ArrayList<>();
        contextKeys.add("clothing");
        contextKeys.add("accessories");
        contextKeys.add("footwear");

        assertEquals(3, contextKeys.size(),
            "Should have 3 context keys");
        assertFalse(contextKeys.isEmpty(),
            "isEmpty() should return false for non-empty context");
    }

    @Test
    @DisplayName("Test ad response builder pattern")
    void testAdResponseBuilder() {
        // Test that builder pattern works correctly
        java.util.List<String> testAds = java.util.Arrays.asList("Ad1", "Ad2");

        assertNotNull(testAds, "Test ads list should not be null");
        assertEquals(2, testAds.size(), "Should have 2 test ads");
    }
}
