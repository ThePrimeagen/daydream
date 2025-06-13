use std::io::{self, BufRead};
use tokio::{task, time::{sleep, Duration}};

async fn write_output() {
    for i in 0..10000 {
        println!("Hello, world!: {}", i);
        sleep(Duration::from_secs(1)).await;
    }
}

async fn read_input() {
    let stdin = io::stdin();
    let reader = io::BufReader::new(stdin);
    for line in reader.lines() {
        match line {
            Ok(line) => println!("{}", line),
            Err(_) => break,
        }
    }
    println!("STDIN closed");
}

#[tokio::main]
async fn main() {
    let write_handle = task::spawn(write_output());
    let read_handle = task::spawn(read_input());

    let _ = tokio::join!(write_handle, read_handle);
}}
