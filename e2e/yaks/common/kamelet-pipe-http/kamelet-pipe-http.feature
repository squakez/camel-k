Feature: Camel K can bind Kamelets via HTTP

  Background:
    Given Camel K resource polling configuration
      | maxAttempts          | 40   |
      | delayBetweenAttempts | 3000 |

  Scenario: Pipe to a HTTP URI should use CloudEvents
    Given Camel K integration display is running
    Then Camel K integration display should print type: org.apache.camel.event
    Then Camel K integration display should print Hello

  Scenario: Remove resources
    Given delete Camel K integration display
    Given delete Pipe timer-source-pipe
