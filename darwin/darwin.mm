// Darwin.mm - Serves as a native objective C wrapper to access MacOS NSFileManager and other MacOS native APIs

#import <Foundation/Foundation.h>

// Checks if a file exists and if true trashes the file
bool MoveToTrash(const char *path) {
    @autoreleasepool {
        NSString *filePath = [NSString stringWithUTF8String:path];
        NSURL *fileURL = [NSURL fileURLWithPath:filePath];
        NSError *error = nil;

        BOOL success = [[NSFileManager defaultManager]
                        trashItemAtURL:fileURL
                        resultingItemURL:nil
                        error:&error];

        if (!success) {
            const char *errorMsg = [[error localizedDescription] UTF8String];
            fprintf(stderr, "[rmapp] Failed to move %s to Trash\n", path);
        }

        return success;
    }
}

// Gets the actual disk usage size of a file
uint64_t GetDiskUsageAtPath(const char *path) {
    @autoreleasepool {
        NSString *pathStr = [NSString stringWithUTF8String:path];
        NSURL *url = [NSURL fileURLWithPath:pathStr];

        // Check if it's a file
        NSNumber *isDirectory = nil;
        NSError *checkError = nil;
        [url getResourceValue:&isDirectory forKey:NSURLIsDirectoryKey error:&checkError];

        if (checkError) {
            NSLog(@"[rmapp] Failed to check if directory: %@ - %@", pathStr, checkError.localizedDescription);
            return 0;
        }

        // If not a directory, just return the file size
        if (![isDirectory boolValue]) {
            NSNumber *fileSize = nil;
            NSError *sizeError = nil;
            [url getResourceValue:&fileSize forKey:NSURLTotalFileAllocatedSizeKey error:&sizeError];
            if (sizeError || fileSize == nil) {
                NSLog(@"[rmapp] Could not retrieve file size for %@: %@", pathStr, sizeError.localizedDescription);
                return 0;
            }
            return [fileSize unsignedLongLongValue];
        }

        // Else, directory: walk it recursively
        NSFileManager *fm = [NSFileManager defaultManager];
        NSDirectoryEnumerator *enumerator = [fm enumeratorAtURL:url
                                     includingPropertiesForKeys:@[NSURLTotalFileAllocatedSizeKey]
                                                        options:NSDirectoryEnumerationSkipsHiddenFiles
                                                   errorHandler:^BOOL(NSURL *url, NSError *error) {
            NSLog(@"[rmapp] Could not retrieve disk usage for %@: %@", url.path, error.localizedDescription);
            return YES;
        }];

        uint64_t totalSize = 0;
        for (NSURL *fileURL in enumerator) {
            NSNumber *fileSize = nil;
            NSError *error = nil;
            [fileURL getResourceValue:&fileSize forKey:NSURLTotalFileAllocatedSizeKey error:&error];
            if (fileSize != nil) {
                totalSize += [fileSize unsignedLongLongValue];
            }
        }

        return totalSize;
    }
}