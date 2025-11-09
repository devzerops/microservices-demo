package hipstershop;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import hipstershop.Demo.Ad;
import hipstershop.Demo.AdRequest;
import hipstershop.Demo.AdResponse;
import io.grpc.stub.StreamObserver;
import io.grpc.testing.GrpcCleanupRule;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.util.Collection;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.mockito.ArgumentCaptor;

/** Unit tests for AdService. */
public class AdServiceTest {

  private Object adServiceInstance;
  private Method getAdsByCategoryMethod;
  private Method getRandomAdsMethod;

  @BeforeEach
  public void setUp() throws Exception {
    // Access the singleton instance via reflection
    Class<?> adServiceClass = Class.forName("hipstershop.AdService");
    Method getInstanceMethod = adServiceClass.getDeclaredMethod("getInstance");
    getInstanceMethod.setAccessible(true);
    adServiceInstance = getInstanceMethod.invoke(null);

    // Get private methods for testing
    getAdsByCategoryMethod = adServiceClass.getDeclaredMethod("getAdsByCategory", String.class);
    getAdsByCategoryMethod.setAccessible(true);

    getRandomAdsMethod = adServiceClass.getDeclaredMethod("getRandomAds");
    getRandomAdsMethod.setAccessible(true);
  }

  @Test
  @DisplayName("Should return ads for valid category 'clothing'")
  public void testGetAdsByCategory_Clothing() throws Exception {
    @SuppressWarnings("unchecked")
    Collection<Ad> ads = (Collection<Ad>) getAdsByCategoryMethod.invoke(adServiceInstance, "clothing");

    assertNotNull(ads, "Ads collection should not be null");
    assertFalse(ads.isEmpty(), "Ads collection should not be empty for 'clothing' category");

    // Verify the ad content
    Ad firstAd = ads.iterator().next();
    assertTrue(firstAd.getText().contains("Tank top"), "Ad text should mention 'Tank top'");
    assertEquals("/product/66VCHSJNUP", firstAd.getRedirectUrl(), "Redirect URL should match");
  }

  @Test
  @DisplayName("Should return ads for valid category 'accessories'")
  public void testGetAdsByCategory_Accessories() throws Exception {
    @SuppressWarnings("unchecked")
    Collection<Ad> ads = (Collection<Ad>) getAdsByCategoryMethod.invoke(adServiceInstance, "accessories");

    assertNotNull(ads, "Ads collection should not be null");
    assertFalse(ads.isEmpty(), "Ads collection should not be empty for 'accessories' category");

    Ad firstAd = ads.iterator().next();
    assertTrue(firstAd.getText().contains("Watch"), "Ad text should mention 'Watch'");
  }

  @Test
  @DisplayName("Should return ads for valid category 'kitchen'")
  public void testGetAdsByCategory_Kitchen() throws Exception {
    @SuppressWarnings("unchecked")
    Collection<Ad> ads = (Collection<Ad>) getAdsByCategoryMethod.invoke(adServiceInstance, "kitchen");

    assertNotNull(ads, "Ads collection should not be null");
    assertFalse(ads.isEmpty(), "Ads collection should not be empty for 'kitchen' category");
    assertEquals(2, ads.size(), "Kitchen category should have 2 ads");
  }

  @Test
  @DisplayName("Should return empty collection for invalid category")
  public void testGetAdsByCategory_InvalidCategory() throws Exception {
    @SuppressWarnings("unchecked")
    Collection<Ad> ads = (Collection<Ad>) getAdsByCategoryMethod.invoke(adServiceInstance, "nonexistent");

    assertNotNull(ads, "Ads collection should not be null");
    assertTrue(ads.isEmpty(), "Ads collection should be empty for invalid category");
  }

  @Test
  @DisplayName("Should return random ads with correct count")
  public void testGetRandomAds() throws Exception {
    @SuppressWarnings("unchecked")
    List<Ad> ads = (List<Ad>) getRandomAdsMethod.invoke(adServiceInstance);

    assertNotNull(ads, "Random ads list should not be null");
    assertEquals(2, ads.size(), "Should return exactly 2 random ads (MAX_ADS_TO_SERVE)");

    for (Ad ad : ads) {
      assertNotNull(ad.getText(), "Ad text should not be null");
      assertNotNull(ad.getRedirectUrl(), "Ad redirect URL should not be null");
      assertFalse(ad.getText().isEmpty(), "Ad text should not be empty");
      assertFalse(ad.getRedirectUrl().isEmpty(), "Ad redirect URL should not be empty");
    }
  }

  @Test
  @DisplayName("Should handle gRPC getAds request with context keys")
  public void testGetAds_WithContextKeys() throws Exception {
    // Create AdServiceImpl instance
    Class<?> adServiceClass = Class.forName("hipstershop.AdService");
    Class<?>[] innerClasses = adServiceClass.getDeclaredClasses();
    Class<?> adServiceImplClass = null;

    for (Class<?> innerClass : innerClasses) {
      if (innerClass.getSimpleName().equals("AdServiceImpl")) {
        adServiceImplClass = innerClass;
        break;
      }
    }

    assertNotNull(adServiceImplClass, "AdServiceImpl class should be found");
    Object adServiceImpl = adServiceImplClass.getDeclaredConstructor().newInstance();

    // Create request with context keys
    AdRequest request = AdRequest.newBuilder()
        .addContextKeys("clothing")
        .addContextKeys("accessories")
        .build();

    // Mock StreamObserver
    @SuppressWarnings("unchecked")
    StreamObserver<AdResponse> responseObserver = mock(StreamObserver.class);

    // Invoke getAds method
    Method getAdsMethod = adServiceImplClass.getMethod("getAds", AdRequest.class, StreamObserver.class);
    getAdsMethod.invoke(adServiceImpl, request, responseObserver);

    // Verify response
    ArgumentCaptor<AdResponse> responseCaptor = ArgumentCaptor.forClass(AdResponse.class);
    verify(responseObserver).onNext(responseCaptor.capture());
    verify(responseObserver).onCompleted();
    verify(responseObserver, never()).onError(any());

    AdResponse response = responseCaptor.getValue();
    assertNotNull(response, "Response should not be null");
    assertFalse(response.getAdsList().isEmpty(), "Response should contain ads");
    assertTrue(response.getAdsList().size() > 0, "Response should have at least one ad");
  }

  @Test
  @DisplayName("Should return random ads when no context keys provided")
  public void testGetAds_WithoutContextKeys() throws Exception {
    // Create AdServiceImpl instance
    Class<?> adServiceClass = Class.forName("hipstershop.AdService");
    Class<?>[] innerClasses = adServiceClass.getDeclaredClasses();
    Class<?> adServiceImplClass = null;

    for (Class<?> innerClass : innerClasses) {
      if (innerClass.getSimpleName().equals("AdServiceImpl")) {
        adServiceImplClass = innerClass;
        break;
      }
    }

    assertNotNull(adServiceImplClass, "AdServiceImpl class should be found");
    Object adServiceImpl = adServiceImplClass.getDeclaredConstructor().newInstance();

    // Create request without context keys
    AdRequest request = AdRequest.newBuilder().build();

    // Mock StreamObserver
    @SuppressWarnings("unchecked")
    StreamObserver<AdResponse> responseObserver = mock(StreamObserver.class);

    // Invoke getAds method
    Method getAdsMethod = adServiceImplClass.getMethod("getAds", AdRequest.class, StreamObserver.class);
    getAdsMethod.invoke(adServiceImpl, request, responseObserver);

    // Verify response
    ArgumentCaptor<AdResponse> responseCaptor = ArgumentCaptor.forClass(AdResponse.class);
    verify(responseObserver).onNext(responseCaptor.capture());
    verify(responseObserver).onCompleted();
    verify(responseObserver, never()).onError(any());

    AdResponse response = responseCaptor.getValue();
    assertNotNull(response, "Response should not be null");
    assertEquals(2, response.getAdsList().size(), "Should return 2 random ads");
  }

  @Test
  @DisplayName("Should verify all product categories have ads")
  public void testAllCategories() throws Exception {
    String[] categories = {"clothing", "accessories", "footwear", "hair", "decor", "kitchen"};

    for (String category : categories) {
      @SuppressWarnings("unchecked")
      Collection<Ad> ads = (Collection<Ad>) getAdsByCategoryMethod.invoke(adServiceInstance, category);

      assertNotNull(ads, "Ads for category '" + category + "' should not be null");
      assertFalse(ads.isEmpty(), "Category '" + category + "' should have at least one ad");
    }
  }

  @Test
  @DisplayName("Should verify ad structure is valid")
  public void testAdStructure() throws Exception {
    @SuppressWarnings("unchecked")
    Collection<Ad> ads = (Collection<Ad>) getAdsByCategoryMethod.invoke(adServiceInstance, "clothing");

    for (Ad ad : ads) {
      assertNotNull(ad.getText(), "Ad text should not be null");
      assertNotNull(ad.getRedirectUrl(), "Ad redirect URL should not be null");

      assertFalse(ad.getText().isEmpty(), "Ad text should not be empty");
      assertFalse(ad.getRedirectUrl().isEmpty(), "Ad redirect URL should not be empty");

      assertTrue(ad.getRedirectUrl().startsWith("/product/"),
          "Redirect URL should start with /product/");
      assertTrue(ad.getText().contains("for sale") || ad.getText().contains("off") ||
          ad.getText().contains("free"),
          "Ad text should contain promotional language");
    }
  }
}
