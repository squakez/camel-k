package org.acme.myservice;

import org.apache.camel.builder.RouteBuilder;

import org.springframework.stereotype.Component;

@Component
public class HttpService extends RouteBuilder {
  @Override
  public void configure() throws Exception {
    from("platform-http:/hello?httpMethodRestrict=GET")
        .setBody(constant("resource:file:/tmp/app/data/my-file.txt"));
  }
}
