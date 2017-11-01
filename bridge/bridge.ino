#define VERSION 120

// Conditions only for Arduino UNO
#define PIN_RST_N 9
#define PIN_SS 10
#define PIN_MOSI 11
#define PIN_MISO 12
#define PIN_SCK 13

//0 :5V 1:3.3V
#define OUTPUT_power 0
#define BAUD_RATE 57600
#define SERIAL_READ_WAIT_US (1000000 * 10 / (BAUD_RATE) + 1)

#include <SPI.h>

// https://www.arduino.cc/en/Reference/PortManipulation
#define PORTB_MASK_SS (1<<((PIN_SS)-8))
#define PORTB_MASK_RST_N (1<<((PIN_RST_N)-8))

// only for Arduino UNO
void set_ss_pin(int val) {
	if (val == HIGH)
		PORTB |= PORTB_MASK_SS;
	else
		PORTB &= ~PORTB_MASK_SS;
}

// only for Arduino UNO
void set_rst_pin(int val) {
	if (val == HIGH)
		PORTB |= PORTB_MASK_RST_N;
	else
		PORTB &= ~PORTB_MASK_RST_N;
}

void if_write(char addr, unsigned const char* data, int size) {
	// Serial.print("#R");
	// Serial.print(addr, DEC);
	set_ss_pin(LOW);
	SPI.transfer(addr);
	for (int i=0; i<size; i++) {
		// Serial.print(",");
		// Serial.print(data[i], DEC);
		SPI.transfer(data[i]);
	}
	// Serial.println();
	set_ss_pin(HIGH);
}

void if_s_write(char addr, unsigned char data) {
	if_write(addr, &data, 1);
}

unsigned char if_s_read(char addr) {
	unsigned char rcv;
	set_ss_pin(LOW);
	SPI.transfer(0x80|addr);
	rcv = SPI.transfer(0x00);
	set_ss_pin(HIGH);
	return rcv;
}

void init_825(void) {
	// https://github.com/yamaha-webmusic/ymf825board/blob/master/manual/fbd_spec1.md

	set_rst_pin(LOW);
	delay(1);
	set_rst_pin(HIGH);

	if_s_write(29, OUTPUT_power); // Power Rail Selection
	if_s_write(2, 0x0E); // Analog Block Power-down control
	delay(1);
	if_s_write(0, 1); // Clock Enable
	if_s_write(1, 0); // Reset = ALRST

	// Soft Reset
	if_s_write(26, 0xA3);
	delay(1);
	if_s_write(26, 0);
	delay(30);

	// power-down control
	if_s_write(2, 0x04);
	delay(1);
	if_s_write(2, 0);
	
	if_s_write(25, 0x38 << 2); // MASTER VOL [0, 0x3F]
	if_s_write(27, 0x3F);  // interpolation MUTE_ITIME
	if_s_write(20, 0);     // interpolation DIR_MT
	if_s_write(3, 1);      // Analog Gain

	// Sequencer Setting (AllKeyOff, AllMute, AllEGRst, etc)
	if_s_write(8, 0xF6);
	delay(21);
	if_s_write(8, 0);
	if_s_write(9, 0x10 << 3);  // SEQ_Vol [0, 0x1F]
	if_s_write(10, 0);

	// Sequencer Time unit Setting = MS_S
	if_s_write(23, 0x40);
	if_s_write(24, 0x00);
}

void _keyon(unsigned char fnumh, unsigned char fnuml){
	if_s_write( 0x0B, 0x00 );//voice num
	if_s_write( 0x0C, 0x54 );//vovol
	if_s_write( 0x0D, fnumh );//fnum
	if_s_write( 0x0E, fnuml );//fnum
	if_s_write( 0x0F, 0x40 );//keyon = 1  
}
 
void _keyoff(void){
	if_s_write( 0x0F, 0x00 );//keyon = 0
}

void _testplay() {
	unsigned char tone_data[35] ={
		0x81,//header
		//T_ADR 0
		0x01,0x85,
		0x00,0x7F,0xF4,0xBB,0x00,0x10,0x40,
		0x00,0xAF,0xA0,0x0E,0x03,0x10,0x40,
		0x00,0x2F,0xF3,0x9B,0x00,0x20,0x41,
		0x00,0xAF,0xA0,0x0E,0x01,0x10,0x40,
		0x80,0x03,0x81,0x80,
	};
	if_s_write(0x08, 0xF6);
	delay(1);
	if_s_write(0x08, 0x00);
	if_write(0x07, tone_data, 35); //write to FIFO

	if_s_write(0x0F, 0x30); // keyon = 0
	if_s_write(0x10, 0x71); // chvol
	if_s_write(0x11, 0x00); // XVB
	if_s_write(0x12, 0x08); // FRAC
	if_s_write(0x13, 0x00); // FRAC

	_keyon(0x14, 0x65); delay(200); _keyoff(); delay(100);
	_keyon(0x1c, 0x11); delay(200); _keyoff(); delay(100);
	_keyon(0x1c, 0x42); delay(200); _keyoff(); delay(100);
	_keyon(0x1c, 0x5d); delay(200); _keyoff(); delay(100);
	_keyon(0x24, 0x17); delay(200); _keyoff(); delay(100);
}

void setup() {
	Serial.begin(BAUD_RATE, SERIAL_8E1);
	Serial.println("#setup");
	pinMode(PIN_RST_N, OUTPUT);
	pinMode(PIN_SS, OUTPUT);
	set_ss_pin(HIGH);
	
	Serial.println("#init spi");
	SPI.setBitOrder(MSBFIRST);
	SPI.setClockDivider(SPI_CLOCK_DIV8); // 16/8 MHz (on Arduino Nano)
	SPI.setDataMode(SPI_MODE0);
	SPI.begin();
	
	Serial.println("#init ymf825");
	init_825();

	// _testplay();

	Serial.print("version ");
	Serial.println(VERSION, DEC);
	Serial.println("ready");
}

#define BUFSIZE 320
unsigned char buf[BUFSIZE];
int bufHead = 0;
int bufLen = 0;
int readOnDemand = 0;

int read() {
	if (bufLen == 0) {
		while (!Serial.available()) delayMicroseconds(SERIAL_READ_WAIT_US);
		readOnDemand++;
		if (6 <= readOnDemand) {
			Serial.println("=6");
			readOnDemand = 0;
		}
	return Serial.read();
	}
	int v = buf[bufHead++];
	bufHead %= BUFSIZE;
	bufLen--;
	return v;
}

unsigned long waitUntil = 0;
bool first = true;

void loop() {
	if (first || micros() < waitUntil) {
		int read = readOnDemand;
		readOnDemand = 0;
		while (bufLen < BUFSIZE && Serial.available()) {
			buf[(bufHead+bufLen) % BUFSIZE] = Serial.read();
			bufLen++;
			read++;
			if (60 <= read) {
				Serial.print("=");
				Serial.println(read, DEC);
				read = 0;
			}
			if (first) {
				delay(100);
				first = false;
			}
		}
		if (micros() + SERIAL_READ_WAIT_US * 3 < waitUntil) {
			readOnDemand += read;
			return;
		}
		if (0 < read) {
			Serial.print("=");
			Serial.println(read, DEC);
		}
		return;
	}

	int addr = read();
	int size = 1;
	if (addr & 0x80) {
		int h = read();
		int l = read();
		size = h << 8 | l;
	}
	addr &= 0x7F;
	if (addr == 0x7F) {
		if (size == 0xFFFF) {
			// Serial.print("#T");
			while (Serial.available()) Serial.read();
			Serial.end();
			return;
		} else {
			// Serial.print("#W");
			// Serial.println(size, DEC);
			waitUntil = micros() + (unsigned long)size * 1000;
		}
	} else {
		set_ss_pin(LOW);
		SPI.transfer(addr);
		while (0 < size--) SPI.transfer(read());
		set_ss_pin(HIGH);
	}
}
