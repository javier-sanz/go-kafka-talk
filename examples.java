// Examples taken from O'Reilly's book "Kafka the definitve guide"
public class Exammples {
    
    public void SendSyncronousMessage() {
        ProducerRecord<String, String> record =
            new ProducerRecord<>("CustomerCountry", "Precision Products", "France");
        try {
            // Send returns a future
            producer.send(record).get();
        } catch (Exception e) {
            ...
        }
    }

    public void SendAsyncronousMessage() {
        ProducerRecord<String, String> record =
            new ProducerRecord<>("CustomerCountry", "Precision Products", "France");
        try {
            // Send returns a future
            futureResult = producer.send(record);
        } catch (Exception e) {
            ...
        }
        // Handle futureResult as you wish
    }

    private class DemoProducerCallback implements Callback {
        @Override
        public void onCompletion(RecordMetadata recordMetadata, Exception e) {
            if (e != null) {
              e.printStackTrace();
            }
        }
    }

    public void SendAsyncronousMessageWithCallback() {
        ProducerRecord<String, String> record =
            new ProducerRecord<>("CustomerCountry", "Biomedical Materials", "USA");
        producer.send(record, new DemoProducerCallback()); 
    }
        

}