<?php
namespace App\Tests;

use Lcobucci\JWT\Token\Parser;
use Bundles\UserContext\Entity\User;
use Bundles\UserContext\Service\EmailSender;
use Doctrine\ORM\EntityManagerInterface;
use Lcobucci\JWT\Encoding\JoseEncoder;
use Symfony\Bundle\FrameworkBundle\KernelBrowser;
use Symfony\Bundle\FrameworkBundle\Test\WebTestCase;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class RestTestCase extends WebTestCase
{
    public const CLIENT_ID = 'c3ff36379fd0aff317297ed1d1b45b80';

    public const CLIENT_SECRET = '7d094bf4175b0a95890b30a8c260597449b086aac70729444d72a4b2d11f3ee0ba05356ee4e63bd28f26f8f63ae40c685f6e0ae9512b38902c63e652b1c6621c';

    protected static KernelBrowser $client;

    protected ?string $token = null;

    protected array $messages = [];

    protected string $tracker = '';

    protected ?EntityManagerInterface $em;

    protected string $username = 'john';

    protected string $password = '112233';

    protected array $scopes = ['user'];

    protected $bus;

    protected bool $anonymous = false;

    protected bool $overrideUuidSnapshot = false;

    protected bool $fix = false;

    protected array $emails = [];

    protected function services(): void
    {

    }

    protected function mockEmails(): void
    {
        $emailService = $this->createMock(EmailSender::class);
        self::getContainer()->set('test.email_sender', $emailService);

        $emailService
            ->method('send')
            ->willReturnCallback(function (int $templateId, array $params, User $user) {
                $this->emails[] = [$templateId, $params, $user];
            });
    }

    public function setUp(): void
    {
        static::$client = self::createClient([
            'environment' => 'test',
        ]);

        static::$client->disableReboot();

        $this->mockEmails();
        $this->services();
        $this->busMock();
        $this->busHandle();

        $this->em = self::getContainer()->get('doctrine.orm.default_entity_manager');
        $this->em->beginTransaction();

        if ($this->username !== '' && $this->username !== '0') {
            $user = $this->em->getRepository(User::class)->findOneBy([
                'username' => $this->username,
            ]);

            if ($user instanceof User) {
                self::getContainer()->get('test.brute_force_security')->invalidate($user);
            }
        }
    }

    public function tearDown(): void
    {
        $this->em->rollback();

        parent::tearDown();
    }

    public function getUrl(string $name, array $params = []): string
    {
        return self::$container->get('router')->generate($name, $params);
    }

    public function deserialize(Response $response, int $statusCode = Response::HTTP_OK)
    {
        $content = $response->getContent();
        $this->assertEquals($statusCode, $response->getStatusCode(), $content);

        return json_decode($content, true);
    }

    public function apiRequest(string $uri, string $method = Request::METHOD_GET, array $params = [], array $headers = []): Response
    {
        $this->loadToken();

        $headers = array_merge($headers, $this->anonymous ? [] : [
            'HTTP_AUTHORIZATION' => $this->token,
        ]);

        return $this->apiPublicRequest($uri, $method, $params, $headers);
    }

    public function apiPublicRequest(string $uri, string $method = Request::METHOD_GET, array $params = [], array $headers = []): Response
    {
        if ($this->tracker !== '' && $this->tracker !== '0') {
            $headers['HTTP_X-TRACKER-ID'] = $this->tracker;
        }

        static::$client->request($method, $uri, $params, [], $headers);

        return static::$client->getResponse();
    }

    protected function loadToken(): void
    {
        if ($this->token) {
            return;
        }

        $url  = $this->getUrl('oauth2_token', []);
        $body = [
            'client_id'     => self::CLIENT_ID,
            'client_secret' => self::CLIENT_SECRET,
            'grant_type'    => 'password',
            'username'      => $this->username,
            'password'      => $this->password,
            'scope'         => $this->scopes,
        ];

        static::$client->request(Request::METHOD_POST, $url, $body);
        $response = static::$client->getResponse();
        $content  = $this->deserialize($response);

        $this->assertArrayHasKey('access_token', $content);

        $parser    = new Parser(new JoseEncoder());
        $parsedJWT = $parser->parse($content['access_token']);
        $userData  = $parsedJWT->claims()->get('user');

        $this->assertArrayHasKey('roles', $userData);
        $this->assertArrayHasKey('username', $userData);

        $this->token = "{$content['token_type']} {$content['access_token']}";
    }

    /**
     * @param array|string $snapshot
     *
     * @throws \ReflectionException
     */
    protected function assertSnapshot(
        $snapshot,
        bool $fixRandom = false,
        string $ext = 'json',
        string $suffix = '',
        bool $escapeUnicode = false
    ): void {
        if ($fixRandom && 'json' === $ext) {
            $snapshot = $this->assertId($snapshot);
        }

        if ($fixRandom && 'txt' === $ext) {
            $snapshot = preg_replace('/(ID:\s+[0-9]+)/ui', 'ID: <id>', $snapshot);
        }

        $name       = preg_replace('/[^a-z0-9]/ui', '', $this->getName() . $suffix);
        $projectDir = self::$container->getParameter('kernel.project_dir');
        $dataSetDir = "{$projectDir}/tests/_data/{$this->getTestPath()}";

        if (!file_exists($dataSetDir)) {
            mkdir($dataSetDir, 0777, true);
        }

        $filePath = "{$dataSetDir}/{$name}.{$ext}";

        if (!$this->fix && file_exists($filePath)) {
            $fixer = (fn(?string $string) => preg_replace('/(\s|\b|\n)+/ui', '', $string ?: ''));

            $expected = \is_array($snapshot)
                ? json_decode(file_get_contents($filePath), true)
                : $fixer(file_get_contents($filePath));

            $this->assertEquals($expected, \is_array($snapshot) ? $snapshot : $fixer($snapshot));

            return;
        }

        $options = JSON_PRETTY_PRINT;

        if ($escapeUnicode) {
            $options |= JSON_UNESCAPED_UNICODE;
        }

        file_put_contents($filePath, \is_array($snapshot) ? json_encode($snapshot, $options) : $snapshot);
        if (!$this->fix) {
            $this->markTestIncomplete('Snapshot has been created.');
        }
    }

    protected function assertJsonSnapshot(?array $json, string $suffix = '')
    {
        $this->assertSnapshot($json, true, 'json', $suffix, true);
    }

    /**
     * @param array|string $data
     *
     * @return array
     */
    private function assertId($data)
    {
        $original = $data;

        if (!\is_array($data)) {
            $data = explode("\n", $data);
            $data = array_filter($data, fn ($item) => mb_strlen((string) $item) > 0);
        }

        if (0 === \count($data)) {
            return $original;
        }

        $new = [];

        foreach ($data as $key => $item) {
            if (\is_array($item)) {
                $new[$key] = $this->assertId($item);

                continue;
            }

            if ('expires_in' === $key) {
                $new[$key] = '<expires-in>';

                continue;
            }

            if ('access_token' === $key) {
                $new[$key] = '<access-token>';

                continue;
            }

            if (preg_match('/\/([a-z0-9\._-]+(\/)?)+\.[a-z0-9]+/ui', (string) $item)) {
                $new[$key] = preg_replace('/\/((\/)?[a-z0-9\._-]+(\/)?)+/ui', '<file-path>', (string) $item);

                continue;
            }

            if ('refresh_token' === $key) {
                $new[$key] = '<refresh-token>';

                continue;
            }

            if ('uuid' === $key) {
                $new[$key] = '<uuid>';

                continue;
            }

            if ('username' === $key) {
                $new[$key] = '<username>';

                continue;
            }

            if (false !== mb_strpos((string) $key, 'date')) {
                $new[$key] = '<date>';

                continue;
            }

            if ($this->overrideUuidSnapshot && preg_match('/[a-z0-9-]{36}/ui', (string) $item)) {
                $new[$key] = preg_replace('/[a-z0-9-]{36}/ui', '<uuid>', (string) $item);

                continue;
            }

            if ('token' === $key && null !== $item) {
                $new[$key] = '<id>';

                continue;
            }

            $datePattern = '/^20[2-9][0-9]-[0-9]{2}-[0-9]{2}(T|\s+)?[0-9]{2}:[0-9]{2}(:[0-9]{2})?(\+[0-9]{2}:[0-9]{2})?(\.[0-9]+Z?)?$/ui';

            if (preg_match($datePattern, (string) $item)) {
                $new[$key] = '<date>';

                continue;
            }

            if (
            preg_match(
                '/^(?P<first>http(s)?:\/\/[a-z0-9-\.]+(net|ru|com|svt|test))(?P<second>.*)$/ui',
                (string) $item,
                $matches
            )
            ) {
                $new[$key] = "host{$matches['second']}";

                continue;
            }

            if (!preg_match('/^([a-zA-Z]+Id|id)$/u', (string) $key) && !preg_match('/^[a-z0-9-]{36}$/ui', (string) $item)) {
                $new[$key] = $item;

                continue;
            }

            if (null === $item) {
                $new[$key] = $item;

                continue;
            }

            if (!\is_int($item)) {
                $this->assertMatchesRegularExpression('/^[a-z0-9- ]+$/ui', $item);

                $new[$key] = preg_match('/^[a-z0-9-]{36}$/ui', (string) $key) ? '<id>' : $item;

                continue;
            }

            $this->assertIsInt($item);

            $new[$key] = '<id>';
        }

        return $new;
    }

    /**
     * @throws \ReflectionException
     */
    protected function getTestPath(): string
    {
        $reflection = new \ReflectionClass(static::class);

        if (false === preg_match('/tests\/(?P<path>.*)\.php$/ui', $reflection->getFileName(), $matches)) {
            throw new \RuntimeException("Can't get path!");
        }

        return $matches['path'];
    }

    protected function busMock(): void
    {
    }

    protected function busHandle(): void
    {
    }

    protected function sqlProfiled(\Closure $closure, int $expectedCount, string $mode = 'assertEquals')
    {
        static::$client->enableProfiler();
        $profiler  = self::$container->get('profiler');

        try {
            $req = $this->createMock(Request::class);
            $res = $this->createMock(Response::class);

            $profiler->get('db')->collect($req, $res);
            $profiler->get('db')->reset();

            $sqlBefore = $profiler->get('db')->getQueryCount();
        } catch (\Exception $exception) {
            unset($exception);

            $sqlBefore = 0;
        }

        $result   = $closure();
        $profiler->get('db')->getQueryCount();

        $offset = 0;

        if ($offset !== 0) {
            $offset = $sqlBefore - 1;
        }

        $requests = \array_slice($profiler->get('db')->getQueries()['default'], $offset);
        $slowest  = $requests;
        usort($slowest, fn ($a, $b) => $a['executionMS'] > $b['executionMS'] ? -1 : 1);

        $this->{$mode}(
            $expectedCount,
            \count($requests),
            print_r([
                'requests' => $requests,
                'slowest'  => $slowest,
            ], true)
        );

        return $result;
    }
}
