//go:build darwin && cgo

#import <Foundation/Foundation.h>
#import <Speech/Speech.h>
#import <AVFoundation/AVFoundation.h>
#include <stdlib.h>
#include <string.h>
#include "apple_native.h"

// Manual reference counting: Go owns one reference to the session. Copied
// dispatch/authorization/result blocks retain it independently. All mutable
// state (including task creation and cancellation) is guarded by @synchronized.
@interface SkyEyeSpeechSession : NSObject {
@public
    BOOL done;
    NSString *text;
    NSString *failure;
    NSData *audio;
    NSArray *hints;
    double sampleRate;
    SFSpeechRecognizer *recognizer;
    SFSpeechAudioBufferRecognitionRequest *request;
    SFSpeechRecognitionTask *task;
}
- (void)finish:(NSString *)result error:(NSString *)error;
- (void)start;
@end

static NSString *configurationError(void) {
    id description = [[NSBundle mainBundle] objectForInfoDictionaryKey:@"NSSpeechRecognitionUsageDescription"];
    if (![description isKindOfClass:[NSString class]] ||
        [[description stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceAndNewlineCharacterSet]] length] == 0) {
        return @"Apple Speech requires a nonempty NSSpeechRecognitionUsageDescription in the executable's embedded Info.plist; embed it before requesting Speech permission.";
    }
    if (@available(macOS 10.15, *)) {
        SFSpeechRecognizer *r = [[SFSpeechRecognizer alloc] initWithLocale:[NSLocale localeWithLocaleIdentifier:@"en-US"]];
        BOOL supported = r && r.supportsOnDeviceRecognition;
        [r release];
        if (!supported) return @"Apple Speech on-device en-US recognition is unavailable. Install English speech resources in macOS settings; server recognition is never used.";
    } else {
        return @"Apple Speech requires macOS 10.15 or newer for on-device recognition.";
    }
    return nil;
}

@implementation SkyEyeSpeechSession
- (void)finish:(NSString *)result error:(NSString *)error {
    @synchronized(self) {
        if (done) return;
        done = YES;
        text = [result copy];
        failure = [error copy];
        // Break the session -> task -> result-block -> session cycle even if
        // cancellation never delivers another callback.
        SFSpeechRecognitionTask *oldTask = task;
        task = nil;
        [oldTask cancel];
        [oldTask release];
    }
}
- (void)start {
    @autoreleasepool {
        @synchronized(self) {
            if (done) return;
            if (@available(macOS 10.15, *)) {
                recognizer = [[SFSpeechRecognizer alloc] initWithLocale:[NSLocale localeWithLocaleIdentifier:@"en-US"]];
                if (!recognizer.supportsOnDeviceRecognition || !recognizer.available) {
                    [self finish:nil error:@"Apple Speech local en-US recognizer is unavailable; check installed speech resources."];
                    return;
                }
                // Never use the main operation queue: the host may not run a
                // Cocoa main run loop and Go must not block that queue.
                NSOperationQueue *queue = [[NSOperationQueue alloc] init];
                queue.maxConcurrentOperationCount = 1;
                recognizer.queue = queue;
                [queue release];
                request = [[SFSpeechAudioBufferRecognitionRequest alloc] init];
                request.requiresOnDeviceRecognition = YES;
                request.shouldReportPartialResults = NO;
                request.contextualStrings = hints;
                AVAudioFormat *format = [[AVAudioFormat alloc] initWithCommonFormat:AVAudioPCMFormatFloat32 sampleRate:sampleRate channels:1 interleaved:NO];
                AVAudioPCMBuffer *buffer = [[AVAudioPCMBuffer alloc] initWithPCMFormat:format frameCapacity:(AVAudioFrameCount)(audio.length / sizeof(float))];
                [format release];
                if (!buffer || !buffer.floatChannelData) {
                    [buffer release];
                    [self finish:nil error:@"Apple Speech could not allocate the mono Float32 audio buffer."];
                    return;
                }
                buffer.frameLength = buffer.frameCapacity;
                memcpy(buffer.floatChannelData[0], audio.bytes, audio.length);
                task = [[recognizer recognitionTaskWithRequest:request resultHandler:^(SFSpeechRecognitionResult *result, NSError *error) {
                    @autoreleasepool {
                        if (error) [self finish:nil error:[NSString stringWithFormat:@"Apple Speech: %@", error.localizedDescription]];
                        else if (result.isFinal) [self finish:result.bestTranscription.formattedString error:nil];
                    }
                }] retain];
                [request appendAudioPCMBuffer:buffer];
                [request endAudio];
                [buffer release];
            } else {
                [self finish:nil error:@"Apple Speech requires macOS 10.15 or newer."];
            }
        }
    }
}
- (void)dealloc {
    [task release];
    [request release];
    [recognizer release];
    [audio release];
    [hints release];
    [text release];
    [failure release];
    [super dealloc];
}
@end

char *apple_preflight(void) {
    @autoreleasepool {
        NSString *error = configurationError();
        return error ? strdup(error.UTF8String) : NULL;
    }
}

void *apple_begin(const float *pcm, size_t count, double rate, const char *json) {
    @autoreleasepool {
        SkyEyeSpeechSession *session = [[SkyEyeSpeechSession alloc] init];
        session->audio = [[NSData alloc] initWithBytes:pcm length:count * sizeof(float)];
        NSData *data = [[NSString stringWithUTF8String:json] dataUsingEncoding:NSUTF8StringEncoding];
        session->hints = [[NSJSONSerialization JSONObjectWithData:data options:0 error:NULL] retain];
        session->sampleRate = rate;
        dispatch_async(dispatch_get_global_queue(QOS_CLASS_USER_INITIATED, 0), ^{
            @autoreleasepool {
                @synchronized(session) {
                    if (session->done) return;
                    NSString *error = configurationError();
                    if (error) { [session finish:nil error:error]; return; }
                    void (^authorized)(SFSpeechRecognizerAuthorizationStatus) = ^(SFSpeechRecognizerAuthorizationStatus status) {
                        @autoreleasepool {
                            if (status == SFSpeechRecognizerAuthorizationStatusAuthorized) [session start];
                            else [session finish:nil error:@"Apple Speech permission denied or restricted. Enable Speech Recognition for this executable or terminal in System Settings > Privacy & Security > Speech Recognition."];
                        }
                    };
                    SFSpeechRecognizerAuthorizationStatus status = [SFSpeechRecognizer authorizationStatus];
                    if (status == SFSpeechRecognizerAuthorizationStatusNotDetermined) {
                        // No semaphore wait and no dispatch_sync to the main queue.
                        [SFSpeechRecognizer requestAuthorization:authorized];
                    } else {
                        authorized(status);
                    }
                }
            }
        });
        return session;
    }
}

int apple_poll(void *handle, char **text, char **error) {
    @autoreleasepool {
        SkyEyeSpeechSession *session = (SkyEyeSpeechSession *)handle;
        @synchronized(session) {
            if (!session->done) return 0;
            *text = session->text ? strdup(session->text.UTF8String) : NULL;
            *error = session->failure ? strdup(session->failure.UTF8String) : NULL;
            return 1;
        }
    }
}

void apple_release(void *handle) {
    @autoreleasepool {
        SkyEyeSpeechSession *session = (SkyEyeSpeechSession *)handle;
        [session finish:nil error:@"Apple Speech recognition canceled."];
        [session release];
    }
}
