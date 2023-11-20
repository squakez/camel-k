import org.apache.camel.builder.RouteBuilder;

public class HttpService extends RouteBuilder {
  @Override
  public void configure() throws Exception {
    from("platform-http:/hello?httpMethodRestrict=GET")
        .setBody(constant("resource:file:/tmp/app/data/my-file.txt"));
  }
}