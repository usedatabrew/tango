## <p align="center">Tango - Open Source Golang Pipeline with Backpressure</p>

<p align="center">
  <img src="./image/logo.png" alt="">
</p>

Tango can help you implement multi-stage pipeline that will apply backpressure on the data producer channel (will pause
reading)
in case the producer is faster than stage processors. It may be a great choice if you want to create ETL pipeline with
custom processing