#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include <sys/un.h>
#include <sys/socket.h>
#include <sys/types.h>

#define MAX_NETSTRING 1024*1024*1024

// globals
static FILE* sockf = NULL;

void die(const char* msg) __attribute__ ((noreturn));
void setup (void) __attribute__ ((constructor));
const unsigned char* afl_postprocess(const unsigned char* in_buf, unsigned int* inout_len);

void 
die(const char* msg) 
{
  fprintf(stderr, "[SHIM] Fatal error: %s\n", msg);
  exit(-1);
}

// set up our socket on load
void 
setup(void) 
{ 

  int sock;
  socklen_t saddr_len;
  struct sockaddr_un saddr;
  const char* sock_fname;

  if ((sock = socket(AF_UNIX, SOCK_STREAM, 0)) < 0) {
    die("failed to create socket");
  }

  sock_fname = getenv("AFL_FIX_SOCK");
  if (!sock_fname) {
    die("couldn't find AFL_FIX_SOCK in ENV\n");
  }

  memset(&saddr, 0, sizeof(struct sockaddr_un));
  saddr.sun_family = AF_UNIX;
  strlcpy(saddr.sun_path, sock_fname, sizeof(saddr.sun_path));
  saddr_len = (socklen_t)SUN_LEN(&saddr);
  if (connect(sock, (struct sockaddr*)&saddr, saddr_len) < 0) {
    die("error in connect\n - is the listener \
running?\n - does the sockfile have correct permissions?");
  }

  // wrap the fd as a FILE* so we can use stdio functions on it
  // this var is available to afl_postprocess as a global
  sockf = fdopen(sock, "w+");
  if (!sockf) {
    die("failed to fdopen");
  }

}

const unsigned char* 
afl_postprocess(const unsigned char* in_buf, unsigned int* inout_len)
{

  static unsigned char* sock_buf = NULL;
  static size_t sock_buf_sz = 0;
  static unsigned char scratch; // one byte buffer for : and , tokens
  static unsigned int recv_len;

  // send netstring directly from input buffer
  if (*inout_len == 0 || *inout_len > MAX_NETSTRING) {
    die("input > 1GB");
  }
  if (fprintf(sockf, "%u:", *inout_len) < 0) {
    die("failed to write to socket");
  }
  if (fwrite(in_buf, 1, *inout_len, sockf) < *inout_len) {
    die("failed to write to socket");
  }
  if (fwrite(",", 1, 1, sockf) < 1) {
    die("failed to write to socket");
  }

  // read back the (possibly) modified test
  if (fscanf(sockf, "%u", &recv_len) < 1) {
    die("invalid netstring");
  }
  if (recv_len == 0 || recv_len > MAX_NETSTRING) {
    die("invalid netstring");
  }

  // increase the size of sock_buf if neccessary
  if (recv_len >  sock_buf_sz) {
    free(sock_buf); // safe because it's initialized to NULL
    sock_buf = (unsigned char*)calloc((size_t)recv_len, 1);
    if (!sock_buf) {
      die("failed to calloc for recv buffer");
    }
    sock_buf_sz = (size_t)recv_len;
  } else {
    memset(sock_buf, 0, sock_buf_sz);
  }

  // verify separator (':')
  if (fread(&scratch, 1, 1, sockf) < 1) {
    die("failed to read from socket");
  }
  if (scratch != ':') {
    die("invalid netstring");
  }
 
  // read the data
  if (fread(sock_buf, 1, recv_len, sockf) < recv_len) {
    die("failed to read from socket");
  }
  
  // verify terminator (',')
  if (fread(&scratch, 1, 1, sockf) < 1) {
    die("failed to read from socket");
  }
  if (scratch != ',') {
    die("invalid netstring");
  }

  // provide the buffer length to the caller
  *inout_len = recv_len;
  // return the fixed buffer
  return sock_buf;

}
