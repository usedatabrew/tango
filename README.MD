## <p align="center">Tango - Open Source Golang Pipeline with Backpressure (WIP)</p>

<p align="center">
  <img src="./image/logo.png" alt="">
</p>

Tango can help you implement a multi-stage pipeline that will apply backpressure on the data producer channel (will pause
reading)
in case the producer is faster than the stage processors. It may be a great choice if you want to create an ETL pipeline with
custom processing
